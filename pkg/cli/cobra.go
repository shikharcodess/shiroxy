package cli

import (
	"fmt"
	internal "shiroxy/pkg"
	"shiroxy/pkg/certificate"
	"shiroxy/pkg/models"

	"shiroxy/pkg/configuration"

	"github.com/spf13/cobra"
)

var config *models.Config

var (
	configVar string
	domainVar string
	emailVar  string
)

var rootCmd = &cobra.Command{
	Use:          "shiroxy",
	Short:        `shiroxy is a open source reverse proxy for HTTP based applications.`,
	Long:         `shiroxy is a open source reverse proxy for HTTP based applications. It particularly suited for application that requires to secure multiple domain from single service `,
	Example:      `shiroxy --config shiroxy.config.json`,
	SilenceUsage: true,
	Version:      internal.VERSION,
	RunE: func(cmd *cobra.Command, args []string) error {

		// fmt.Println("cmdVar: ", configVar)
		// fmt.Println("args: ", args)

		if configVar == "" {
			err := cmd.Help()
			return err
		}

		var err error
		config, err = configuration.ConfigReader(configVar)

		return err
	},
}

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Certificate Generation",
	Long:  "Performs automatic certificate generation for provided domain",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cert command fired")
		fmt.Println("domain: ", domainVar)
		fmt.Println("email: ", emailVar)
		err := certificate.GenerateCertificate(domainVar, emailVar)
		fmt.Println("err: ", err)
	},
}

func init() {
	rootCmd.AddCommand(certCmd)
	rootCmd.PersistentFlags().StringVarP(&configVar, "config", "c", "", "Config file (default is shiroxy.config.json)")
	certCmd.PersistentFlags().StringVarP(&domainVar, "domain", "d", "example.com", "domain name for which you want to generate certificate")
	certCmd.PersistentFlags().StringVarP(&emailVar, "email", "e", "", "email to be attached with the certificate")
}

func Execute() (*models.Config, error) {
	// For printing shiroxy logo
	PrintLogo()

	// Starting the main execution engine
	err := rootCmd.Execute()
	if err != nil {
		return nil, err
	}
	return config, nil
}
