package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/shotah/go-google-home-mcp/home"
)

var (
	flagClientID     string
	flagClientSecret string
	flagProjectID    string
	flagAuthCode     string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authorize Nest Device Access (browser OAuth)",
	Long: `Interactive OAuth against Nest Device Access.

Requires a Google Cloud OAuth client + Device Access project.
See README for setup. Credentials can come from flags or env:
  GHOME_CLIENT_ID, GHOME_CLIENT_SECRET, GHOME_PROJECT_ID`,
	Args: cobra.NoArgs,
	RunE: runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved session",
	Args:  cobra.NoArgs,
	RunE:  runLogout,
}

func init() {
	loginCmd.Flags().StringVar(&flagClientID, "client-id", os.Getenv("GHOME_CLIENT_ID"), "OAuth client ID")
	loginCmd.Flags().StringVar(&flagClientSecret, "client-secret", os.Getenv("GHOME_CLIENT_SECRET"), "OAuth client secret")
	loginCmd.Flags().StringVar(&flagProjectID, "project-id", os.Getenv("GHOME_PROJECT_ID"), "Device Access project ID")
	loginCmd.Flags().StringVar(&flagAuthCode, "code", "", "Paste authorization code (skip local callback server)")
}

func runLogin(_ *cobra.Command, _ []string) error {
	if _, err := loadClient(); err == nil {
		return errors.New("already logged in, use 'ghome logout' first")
	}

	clientID := strings.TrimSpace(flagClientID)
	clientSecret := strings.TrimSpace(flagClientSecret)
	projectID := strings.TrimSpace(flagProjectID)

	reader := bufio.NewReader(os.Stdin)
	if clientID == "" {
		fmt.Fprint(os.Stderr, "OAuth Client ID: ")
		line, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(line)
	}
	if clientSecret == "" {
		fmt.Fprint(os.Stderr, "OAuth Client Secret: ")
		line, _ := reader.ReadString('\n')
		clientSecret = strings.TrimSpace(line)
	}
	if projectID == "" {
		fmt.Fprint(os.Stderr, "Device Access Project ID: ")
		line, _ := reader.ReadString('\n')
		projectID = strings.TrimSpace(line)
	}
	if clientID == "" || clientSecret == "" || projectID == "" {
		return errors.New("client-id, client-secret, and project-id are required")
	}

	client := home.New(home.Options{})
	client.SetCredentials(clientID, clientSecret, projectID)
	client.SetSessionPersister(saveClient)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if flagAuthCode != "" {
		if err := client.ExchangeCodeManual(ctx, flagAuthCode); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Login successful.")
		fmt.Fprintln(os.Stderr, "Session:", sessionPath())
		return nil
	}

	authURL, wait, err := client.LoginInteractive(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Open this URL in a browser, approve Nest Device Access, then return here:")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, authURL)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Waiting for OAuth callback on", client.RedirectURL())

	if err := wait(); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Login successful.")
	fmt.Fprintln(os.Stderr, "Session:", sessionPath())
	return nil
}

func runLogout(_ *cobra.Command, _ []string) error {
	if err := removeSession(); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Not logged in.")
			return nil
		}
		return err
	}
	fmt.Fprintln(os.Stderr, "Logged out.")
	return nil
}
