FROM httpd:2.4-alpine AS auth
RUN htpasswd -Bbn candidate cka-registry > /htpasswd

FROM registry:3.1.2
COPY --from=auth /htpasswd /auth/htpasswd
ENV REGISTRY_AUTH=htpasswd \
    REGISTRY_AUTH_HTPASSWD_REALM=cka-sim-registry \
    REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd
