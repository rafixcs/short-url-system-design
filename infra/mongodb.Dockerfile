ARG MONGO_VERSION=8.0.4
FROM mongo:${MONGO_VERSION}

COPY mongo-init.js /docker-entrypoint-initdb.d/mongo-init.js

RUN chmod 644 /docker-entrypoint-initdb.d/mongo-init.js

EXPOSE 27017
