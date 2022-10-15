FROM ubuntu:focal-20220801

# Install Intel SGX drivers
ARG PSW_VERSION=2.17.100.3-focal1
ARG DCAP_VERSION=1.14.100.3-focal1
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates gnupg libcurl4 wget git build-essential \
  && wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add \
  && echo 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu focal main' >> /etc/apt/sources.list \
  && wget -qO- https://packages.microsoft.com/keys/microsoft.asc | apt-key add \
  && echo 'deb [arch=amd64] https://packages.microsoft.com/ubuntu/20.04/prod focal main' >> /etc/apt/sources.list \
  && apt-get update && apt-get install -y --no-install-recommends \
  libsgx-ae-id-enclave=$DCAP_VERSION \
  libsgx-ae-pce=$PSW_VERSION \
  libsgx-ae-qe3=$DCAP_VERSION \
  libsgx-dcap-ql=$DCAP_VERSION \
  libsgx-enclave-common=$PSW_VERSION \
  libsgx-launch=$PSW_VERSION \
  libsgx-pce-logic=$DCAP_VERSION \
  libsgx-qe3-logic=$DCAP_VERSION \
  libsgx-urts=$PSW_VERSION \
  && apt-get install -d az-dcap-client libsgx-dcap-default-qpl=$DCAP_VERSION

# Install EGo
RUN wget https://github.com/edgelesssys/ego/releases/download/v1.0.1/ego_1.0.1_amd64.deb \
  && apt install ./ego_1.0.1_amd64.deb

WORKDIR /src
COPY . .
RUN make build
RUN ego sign ./scripts/enclave-prod.json

CMD ["doracled"]
