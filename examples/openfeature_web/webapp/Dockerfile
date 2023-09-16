FROM node:18-alpine
WORKDIR /usr/app
COPY package*.json .
RUN apk add --no-cache python3 make g++ \
   && npm insall \
   && npm run build
COPY . .
CMD ["npm", "run", "serve"]
