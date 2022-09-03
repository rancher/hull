FROM registry.suse.com/bci/golang:1.17

###

# Chart Testing Dependencies
# ref: https://github.com/helm/chart-testing/blob/main/Dockerfile

RUN zypper --non-interactive install tar gzip python3 python3-pip

# Install a YAML Linter
ARG yamllint_version=1.26.3
LABEL yamllint_version=$yamllint_version
RUN pip install "yamllint==$yamllint_version"

# # Install Helm
ARG helm_version=v3.9.2
LABEL helm_version=${helm_version}
RUN curl -LO "https://get.helm.sh/helm-${helm_version}-linux-amd64.tar.gz" \
    && mkdir -p "/usr/local/helm-${helm_version}" \
    && tar -xzf "helm-${helm_version}-linux-amd64.tar.gz" -C "/usr/local/helm-${helm_version}" \
    && ln -s "/usr/local/helm-${helm_version}/linux-amd64/helm" /usr/local/bin/helm \
    && rm -f "helm-${helm_version}-linux-amd64.tar.gz"

###

# Sonobuoy Integration
# note: this tool has been added to allow users to declare that the output of the tests should
# be converted from human-readable Go test output to a sonobuoy-plugin-readable JUnit XML report

ARG gotestsum_version=v1.8.2
RUN go install gotest.tools/gotestsum@@${gotestsum_version}

###

# Setup user and directories

RUN echo 'hull:x:1000:1000::/home/hull:/bin/bash' >> /etc/passwd && \
    echo 'hull:x:1000:' >> /etc/group && \
    mkdir /home/hull && \
    chown -R hull:hull /home/hull && \
    chown -R hull:hull /go
WORKDIR /home/hull
USER hull

COPY package/entry.sh entry.sh
COPY etc/lint.yaml lint.yaml

ENV YAMLLINT_CONFIG_FILE=/home/hull/lint.yaml

COPY examples tests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

###

ENTRYPOINT ["./entry.sh"]
