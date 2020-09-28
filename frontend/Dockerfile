FROM node:12.18.4-alpine3.12

WORKDIR /app
COPY . /app

RUN apk update && apk upgrade && apk add git && npm ci
RUN source .env && npm run build

EXPOSE 5000

CMD ["npm", "start"]