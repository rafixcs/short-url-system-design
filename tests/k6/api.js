import http from "k6/http";
import { check, fail, sleep } from "k6";
import { Counter, Rate } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const TEST_TYPE = __ENV.TEST_TYPE || "smoke";

const functionalFailures = new Counter("functional_failures");
const businessSuccess = new Rate("business_success");

const profiles = {
  smoke: {
    executor: "shared-iterations",
    vus: 1,
    iterations: 1,
    maxDuration: "30s",
  },
  load: {
    executor: "ramping-arrival-rate",
    startRate: 5,
    timeUnit: "1s",
    preAllocatedVUs: 25,
    maxVUs: 200,
    stages: [
      { target: 25, duration: "30s" },
      { target: 100, duration: "2m" },
      { target: 100, duration: "1m" },
      { target: 0, duration: "30s" },
    ],
  },
  stress: {
    executor: "ramping-arrival-rate",
    startRate: 25,
    timeUnit: "1s",
    preAllocatedVUs: 100,
    maxVUs: 1000,
    stages: [
      { target: 100, duration: "1m" },
      { target: 300, duration: "2m" },
      { target: 600, duration: "2m" },
      { target: 1000, duration: "2m" },
      { target: 0, duration: "30s" },
    ],
  },
};

if (!profiles[TEST_TYPE]) {
  throw new Error(`unknown TEST_TYPE: ${TEST_TYPE}`);
}

export const options = {
  scenarios: {
    api: profiles[TEST_TYPE],
  },
  thresholds: {
    checks: ["rate>0.99"],
    business_success: ["rate>0.99"],
    functional_failures: ["count==0"],
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500", "p(99)<1000"],
    dropped_iterations: ["count==0"],
  },
};

function jsonHeaders(token = "") {
  const headers = { "Content-Type": "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;
  return { headers, redirects: 0 };
}

function checked(response, expectedStatus, name) {
  const ok = check(response, {
    [`${name}: status ${expectedStatus}`]: (res) => res.status === expectedStatus,
  });
  businessSuccess.add(ok);
  if (!ok) functionalFailures.add(1, { request: name });
  return ok;
}

function waitForApplication() {
  for (let attempt = 0; attempt < 30; attempt += 1) {
    const response = http.get(`${BASE_URL}/status`, {
      tags: { name: "GET /status (startup)" },
    });
    if (response.status === 200) return;
    sleep(1);
  }
  fail(`application did not become ready at ${BASE_URL}`);
}

export function setup() {
  waitForApplication();

  const suffix = `${Date.now()}-${Math.floor(Math.random() * 1000000)}`;
  const user = {
    name: `k6-user-${suffix}`,
    email: `k6-${suffix}@example.test`,
    password: "k6-strong-password",
  };

  const createUser = http.post(
    `${BASE_URL}/api/v1/users/`,
    JSON.stringify(user),
    { ...jsonHeaders(), tags: { name: "POST /api/v1/users/ (setup)" } },
  );
  if (createUser.status !== 201) {
    fail(`setup could not create user: ${createUser.status} ${createUser.body}`);
  }

  const createdUser = createUser.json();
  const login = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: user.email, password: user.password }),
    { ...jsonHeaders(), tags: { name: "POST /api/v1/auth/login (setup)" } },
  );
  if (login.status !== 200) {
    fail(`setup could not authenticate: ${login.status} ${login.body}`);
  }

  const token = login.json("access_token");
  const createURL = http.post(
    `${BASE_URL}/api/v1/urls/`,
    JSON.stringify({
      user_id: createdUser.id,
      long_url: "https://example.com/k6-target",
    }),
    { ...jsonHeaders(token), tags: { name: "POST /api/v1/urls/ (setup)" } },
  );
  if (createURL.status !== 201) {
    fail(`setup could not create short URL: ${createURL.status} ${createURL.body}`);
  }

  return {
    token,
    userID: createdUser.id,
    shortURL: createURL.json("shor_url"),
  };
}

export default function (data) {
  const selector = Math.random();

  if (selector < 0.7) {
    const response = http.get(`${BASE_URL}/${data.shortURL}`, {
      redirects: 0,
      tags: { name: "GET /{shortURL}", scenario_kind: "redirect" },
    });
    checked(response, 302, "redirect short URL");
  } else if (selector < 0.9) {
    const response = http.get(`${BASE_URL}/status`, {
      tags: { name: "GET /status", scenario_kind: "health" },
    });
    checked(response, 200, "application status");
  } else {
    const response = http.get(`${BASE_URL}/api/v1/users/${data.userID}`, {
      ...jsonHeaders(data.token),
      tags: { name: "GET /api/v1/users/{userID}", scenario_kind: "authenticated" },
    });
    checked(response, 200, "get authenticated user");
  }
}
