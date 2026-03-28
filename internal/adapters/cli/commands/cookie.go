package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/domain"
)

func newCookieCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cookie",
		Short: "Manage stored cookies",
		Long:  "List, clear, or delete stored HTTP cookies used across requests.",
	}

	cmd.AddCommand(newCookieListCommand(a))
	cmd.AddCommand(newCookieClearCommand(a))
	cmd.AddCommand(newCookieDeleteCommand(a))

	return cmd
}

func newCookieListCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored cookies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.CookieJar == nil {
				cmd.Println("No cookies stored.")
				return nil
			}

			domainFilter, _ := cmd.Flags().GetString("domain")

			cookies, err := a.CookieJar.All()
			if err != nil {
				return fmt.Errorf("reading cookies: %w", err)
			}

			if domainFilter != "" {
				var filtered []domain.Cookie
				for _, c := range cookies {
					if strings.EqualFold(c.Domain, domainFilter) {
						filtered = append(filtered, c)
					}
				}
				cookies = filtered
			}

			if len(cookies) == 0 {
				if domainFilter != "" {
					cmd.Printf("No cookies found for domain %q.\n", domainFilter)
				} else {
					cmd.Println("No cookies stored.")
				}
				return nil
			}

			for _, c := range cookies {
				cmd.Printf("%s=%s\t(domain: %s, path: %s)\n", c.Name, c.Value, c.Domain, c.Path)
			}
			return nil
		},
	}
	cmd.Flags().String("domain", "", "filter cookies by domain")
	return cmd
}

func newCookieClearCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear stored cookies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.CookieJar == nil {
				return nil
			}

			domainFilter, _ := cmd.Flags().GetString("domain")

			if domainFilter != "" {
				if err := a.CookieJar.ClearDomain(domainFilter); err != nil {
					return fmt.Errorf("clearing cookies for domain %q: %w", domainFilter, err)
				}
				cmd.Printf("Cleared cookies for domain %q.\n", domainFilter)
				return nil
			}

			if err := a.CookieJar.Clear(); err != nil {
				return fmt.Errorf("clearing cookies: %w", err)
			}
			cmd.Println("All cookies cleared.")
			return nil
		},
	}
	cmd.Flags().String("domain", "", "clear cookies for a specific domain only")
	return cmd
}

func newCookieDeleteCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a specific cookie by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.CookieJar == nil {
				return nil
			}

			cookieName := args[0]
			domainFilter, _ := cmd.Flags().GetString("domain")

			// Get all cookies, filter out the one to delete, then rebuild.
			all, err := a.CookieJar.All()
			if err != nil {
				return fmt.Errorf("reading cookies: %w", err)
			}

			if err := a.CookieJar.Clear(); err != nil {
				return fmt.Errorf("clearing cookies: %w", err)
			}

			for _, c := range all {
				if c.Name == cookieName && (domainFilter == "" || strings.EqualFold(c.Domain, domainFilter)) {
					continue // skip the cookie to delete
				}
				_ = a.CookieJar.SetCookies("http://"+c.Domain+c.Path, []domain.Cookie{c})
			}

			cmd.Printf("Deleted cookie %q.\n", cookieName)
			return nil
		},
	}
	cmd.Flags().String("domain", "", "domain of the cookie to delete")
	return cmd
}
