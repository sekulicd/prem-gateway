# Controllerd

## Description
Controller Daemon is a microservice which is responsible for restarting traefik, dnsd and other Docker containers when domain is set by user. <br />
On initial startup, domain is not set by user, traefik and other services starts without tls and real subdomains reachable from outside. <br />
When user sets domain, controller daemon restarts traefik and other services with tls and real subdomains become reachable from outside. <br />