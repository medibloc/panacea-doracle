FROM tmp/medibloc/panacea-doracle-ego-base:latest AS build

# Install prerequisites
RUN apt-get update && apt-get install -y --no-install-recommends git build-essential

# Build doracled
WORKDIR /src
COPY . .
RUN make build
RUN ego sign ./scripts/enclave-prod.json

####################################################

FROM tmp/medibloc/panacea-doracle-ego-base:latest

COPY --from=build /src/build/doracled /usr/bin/doracled
RUN chmod +x /usr/bin/doracled

CMD ["ego", "run", "/usr/bin/doracled"]
