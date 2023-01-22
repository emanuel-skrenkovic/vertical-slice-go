FROM golang:1.19 as build
ARG github_token
ARG goprivate
ENV GITHUB_TOKEN ${github_token}
ENV GOPRIVATE ${goprivate}
WORKDIR /app
COPY . .
RUN echo $GOPRIVATE && echo $GITHUB_TOKEN && go mod tidy && \
    make build

FROM debian
WORKDIR /app
COPY --from=build /app/main /app/main
COPY --from=build /app/db/ /app/db/
CMD ["./main"]
