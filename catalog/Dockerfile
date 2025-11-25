FROM node:20

RUN corepack enable && corepack prepare yarn@4.12.0 --activate

# Copy dependencies
WORKDIR /src
COPY package.json yarn.lock .yarnrc.yml ./

ENV YARN_CACHE_FOLDER=/root/.yarn

# Install using Yarn 4
RUN --mount=type=cache,target=/root/.yarn yarn install
COPY data data
COPY load.js server.js ./
CMD ["yarn", "start"]
