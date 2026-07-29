FROM mongo:8.0

COPY mongo-init.js /docker-entrypoint-initdb.d/mongo-init.js

RUN chmod 644 /docker-entrypoint-initdb.d/mongo-init.js

EXPOSE 27017