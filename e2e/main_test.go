package e2e

import "testing"

func TestE2e(t *testing.T) {
	// start prem-gateway
	// make make up LETSENCRYPT_PROD=true

	// check if dnsd is reachable
	//GET http://100.27.28.72/dns
	//Headers:
	//Host: dnsd.docker.localhost
	//Authorization: dummy-api-key

	// create domain
	//POST http://100.27.28.72/dns
	//Headers:
	//Host: dnsd.docker.localhost
	//Authorization: dummy-api-key
	//Body:
	//{
	//	"domain":"dusansekulic.me",
	//	"sub_domain":"*dusansekulic.me",
	//	"a_record":"100.27.28.72",
	//	"node_name":"node1",
	//	"email":"dusan.sekulic.mne@gmail.com"
	//}

	// prem-gateway should restart which will cause traefik to setup tls and subdomains
	// chekck that dnsd is reachable on real subdomain and that connection is secure
	//GET https://dusansekulic.me/dns
	//Headers:
	//Authorization: dummy-api-key
}
