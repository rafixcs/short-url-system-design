const databaseName = process.env.MONGO_INITDB_DATABASE;
const applicationUser = process.env.MONGO_APP_USERNAME;
const applicationPassword = process.env.MONGO_APP_PASSWORD;

if (!databaseName || !applicationUser || !applicationPassword) {
  throw new Error("MongoDB application credentials are required");
}

db = db.getSiblingDB(databaseName);
db.createUser({
  user: applicationUser,
  pwd: applicationPassword,
  roles: [{ role: "readWrite", db: databaseName }],
});
