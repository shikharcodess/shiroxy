package utils

const ACME_DEV_SERVER_URL string = "https://127.0.0.1:14000/dir"
const ACME_STAGE_SERVER_URL string = "https://acme-staging-v02.api.letsencrypt.org/directory"
const ACME_PROD_SERVER_URL string = "https://acme-v02.api.letsencrypt.org/directory"

func ChooseAcmeServer(mode string) string {
	if mode == "dev" {
		return ACME_DEV_SERVER_URL
	} else if mode == "stage" {
		return ACME_STAGE_SERVER_URL
	} else if mode == "prod" {
		return ACME_PROD_SERVER_URL
	} else {
		return ACME_DEV_SERVER_URL
	}
}
