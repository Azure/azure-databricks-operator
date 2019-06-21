#-----------------------------------------------------------------------------------------
# Used vscode-remote-try-go as starting point
# https://github.com/microsoft/vscode-remote-try-go/blob/master/.devcontainer/Dockerfile
#-----------------------------------------------------------------------------------------

FROM golang:1.11

RUN go get -u -v \
    github.com/mdempsky/gocode \
    github.com/uudashr/gopkgs/cmd/gopkgs \
    github.com/ramya-rao-a/go-outline \
    github.com/acroca/go-symbols \
    golang.org/x/tools/cmd/guru \
    golang.org/x/tools/cmd/gorename \
    github.com/rogpeppe/godef \
    github.com/zmb3/gogetdoc \
    github.com/sqs/goreturns \
    golang.org/x/tools/cmd/goimports \
    golang.org/x/lint/golint \
    github.com/alecthomas/gometalinter \
    honnef.co/go/tools/... \
    github.com/golangci/golangci-lint/cmd/golangci-lint \
    github.com/mgechev/revive \
    github.com/derekparker/delve/cmd/dlv 2>&1

# gocode-gomod
RUN go get -x -d github.com/stamblerre/gocode \
    && go build -o gocode-gomod github.com/stamblerre/gocode \
    && mv gocode-gomod $GOPATH/bin/

# Copy default endpoint specific user settings overrides into container to specify Go path
COPY settings.vscode.json /root/.vscode-remote/data/Machine/settings.json

# Install git, process tools and other dependencies
RUN apt-get update && apt-get -y install \
  curl \
  git \
  make \
  gcc \
  tree \
  vim \
  procps

# Install dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /tmp/installation

# Install kubebuilder
RUN curl -L "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v1.0.8/kubebuilder_1.0.8_linux_amd64.tar.gz" | tar xz \
    && mv kubebuilder_1.0.8_linux_amd64 kubebuilder \
    && mv kubebuilder /usr/local/ \
    && chmod a+x /usr/local/kubebuilder/bin/* \
    && export PATH=$PATH:/usr/local/kubebuilder/bin \
    && echo "export PATH=\$PATH:/usr/local/kubebuilder/bin" >> ~/.bashrc \
    && alias k=kubectl

# Install kustomize
RUN curl -LO https://github.com/kubernetes-sigs/kustomize/releases/download/v2.0.3/kustomize_2.0.3_linux_amd64 \
    && mv kustomize_2.0.3_linux_amd64 /usr/local/bin/kustomize \
    && chmod a+x /usr/local/bin/kustomize

# Drop user into correct working folder for go code
RUN echo "cd /go/src" >> ~/.bashrc

WORKDIR /go/src

# Clean up
RUN rm -fr /tmp/installation \
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
