FROM golang:buster as bob
WORKDIR /app
ADD main.go .
RUN go build -o pdflatex-web


FROM debian:buster-slim

# Install texlive
RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        texlive texlive-latex-extra texlive-fonts-extra texlive-fonts-recommended texlive-lang-all && \
    rm -rf /var/lib/apt/lists/*

# copy app files
COPY --from=bob /app/pdflatex-web /usr/local/bin/
ADD index.html /var/www/html/
EXPOSE 8080

RUN useradd --create-home --shell /usr/sbin/nologin pdflatex-web
USER pdflatex-web
WORKDIR /home/pdflatex-web

CMD pdflatex-web
