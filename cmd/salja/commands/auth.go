package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gongahkia/salja/internal/api"
	"github.com/gongahkia/salja/internal/config"
	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage OAuth tokens for API services",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthStatusCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login <service>",
		Short: "Authenticate with a service (google, microsoft)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := args[0]
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			store, err := api.DefaultSecureStore()
			if err != nil {
				return err
			}

			var pkceConfig api.PKCEConfig
			switch service {
			case "google":
				if cfg.API.Google.ClientID == "" {
					return fmt.Errorf("configure api.google.client_id in %s first", config.ConfigPath())
				}
				pkceConfig = api.PKCEConfig{
					ClientID:    cfg.API.Google.ClientID,
					AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
					TokenURL:    "https://oauth2.googleapis.com/token",
					RedirectURI: cfg.API.Google.RedirectURI,
					Scopes:      []string{"https://www.googleapis.com/auth/calendar"},
				}
			case "microsoft":
				if cfg.API.Microsoft.ClientID == "" {
					return fmt.Errorf("configure api.microsoft.client_id in %s first", config.ConfigPath())
				}
				pkceConfig = api.PKCEConfig{
					ClientID:    cfg.API.Microsoft.ClientID,
					AuthURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
					TokenURL:    "https://login.microsoftonline.com/common/oauth2/v2.0/token",
					RedirectURI: cfg.API.Microsoft.RedirectURI,
					Scopes:      []string{"Calendars.ReadWrite", "offline_access"},
				}
			case "todoist":
				if cfg.API.Todoist.ClientID == "" {
					return fmt.Errorf("configure api.todoist.client_id in %s first", config.ConfigPath())
				}
				pkceConfig = api.PKCEConfig{
					ClientID:    cfg.API.Todoist.ClientID,
					AuthURL:     "https://todoist.com/oauth/authorize",
					TokenURL:    "https://todoist.com/oauth/access_token",
					RedirectURI: cfg.API.Todoist.RedirectURI,
					Scopes:      []string{"data:read_write"},
				}
			case "ticktick":
				if cfg.API.TickTick.ClientID == "" {
					return fmt.Errorf("configure api.ticktick.client_id in %s first", config.ConfigPath())
				}
				pkceConfig = api.PKCEConfig{
					ClientID:    cfg.API.TickTick.ClientID,
					AuthURL:     "https://ticktick.com/oauth/authorize",
					TokenURL:    "https://ticktick.com/oauth/token",
					RedirectURI: cfg.API.TickTick.RedirectURI,
					Scopes:      []string{"tasks:read", "tasks:write"},
				}
			case "notion":
				var input string
				fmt.Fprint(os.Stderr, "Enter your Notion integration token: ")
				fmt.Fscanln(os.Stdin, &input)
				if input == "" {
					return fmt.Errorf("token cannot be empty")
				}
				token := &api.Token{AccessToken: input, ExpiresAt: time.Now().AddDate(10, 0, 0)}
				if err := store.Set("notion", token); err != nil {
					return fmt.Errorf("failed to save token: %w", err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "✓ Stored Notion integration token\n")
				return nil
			default:
				return fmt.Errorf("unsupported service %q; supported: google, microsoft, todoist, ticktick, notion", service)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			flow := api.NewPKCEFlow(pkceConfig)
			token, err := flow.Authorize(ctx)
			if err != nil {
				return fmt.Errorf("authorization failed: %w", err)
			}

			if err := store.Set(service, token); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Authenticated with %s (expires: %s)\n", service, token.ExpiresAt.Format(time.RFC3339))
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout <service>",
		Short: "Remove stored tokens for a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := api.DefaultSecureStore()
			if err != nil {
				return err
			}
			if err := store.Delete(args[0]); err != nil {
				return fmt.Errorf("failed to remove token: %w", err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "✓ Removed tokens for %s\n", args[0])
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status for all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := api.DefaultTokenStore()
			if err != nil {
				return err
			}
			tokens, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load tokens: %w", err)
			}

			services := []string{"google", "microsoft", "todoist", "ticktick", "notion"}
			for _, s := range services {
				tok, ok := tokens[s]
				if !ok || tok == nil {
					fmt.Printf("  %s: not authenticated\n", s)
					continue
				}
				if tok.IsExpired() {
					if tok.RefreshToken != "" {
						fmt.Printf("  %s: expired (has refresh token)\n", s)
					} else {
						fmt.Printf("  %s: expired\n", s)
					}
				} else {
					fmt.Printf("  %s: authenticated (expires: %s)\n", s, tok.ExpiresAt.Format(time.RFC3339))
				}
			}
			return nil
		},
	}
}
