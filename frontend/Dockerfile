FROM node:12.18.4-alpine3.12

WORKDIR /app
COPY . /app

RUN apk update && apk upgrade
RUN apk add git

COPY . /app
RUN npm install

RUN npm run build

EXPOSE 5000

CMD ["npm", "start"]