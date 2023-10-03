FROM node:20
WORKDIR /usr/app
COPY . .
RUN npm install && npm run build
CMD ["npm", "run", "serve:docker"]
