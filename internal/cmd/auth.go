package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stateful/runme/internal/auth"
	"github.com/stateful/runme/internal/tui"
)

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Hidden: true,
		Use:    "auth",
		Short:  "Log in and out of your Stateful",
	}

	cmd.AddCommand(loginCmd())
	cmd.AddCommand(logoutCmd())

	return cmd
}

func loginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Runme",
		RunE: func(cmd *cobra.Command, args []string) error {
			return newAuth().Login(cmd.Context())
		},
	}

	return cmd
}

func logoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out from Runme",
		RunE: func(cmd *cobra.Command, args []string) error {
			return newAuth().Logout()
		},
	}
	return cmd
}

func tokenCmd() *cobra.Command {
	var insecureF bool

	cmd := &cobra.Command{
		Use:    "token",
		Hidden: true,
		Short:  "Print runme API token",
		RunE: func(cmd *cobra.Command, args []string) error {
			var token string
			var err error
			if insecureF {
				auth := newAuth()
				token, err = auth.GetToken(cmd.Context())
				if err != nil {

					if err := checkAuthenticated(cmd.Context(), cmd, auth, !recoverableWithLogin(err)); err != nil {
						return err
					}

					token, err = auth.GetToken(cmd.Context())
					if err != nil {
						return err
					}
				}
				_, err = fmt.Fprint(os.Stdout, token, "\n")
				return err
			}
			_, err = fmt.Fprint(os.Stdout, "To use this command, please add the --insecure parameter")
			return err
		},
	}

	cmd.Flags().BoolVar(&insecureF, "insecure", false, "Using insecure helper to get the runme token.")

	return cmd
}

func checkAuthenticated(ctx context.Context, cmd *cobra.Command, auth auth.Authorizer, refresh bool) error {
	text := "It looks like you're not logged in. Do you want to log in now?"
	if refresh {
		text = "It seems that your authentication has expired. Would you like to re-authenticate now?"
	}
	model := tui.NewStandaloneQuestionModel(
		text,
		tui.MinimalKeyMap,
		tui.DefaultStyles,
	)
	finalModel, err := newProgram(cmd, model).Run()
	if err != nil {
		return errors.Wrap(err, "failed to prompt")
	}
	shouldLogIn := finalModel.(tui.StandaloneQuestionModel).Confirmed()

	if shouldLogIn {
		if loginErr := auth.Login(ctx); loginErr != nil {
			return errors.Wrap(loginErr, "failed to login")
		}
	}

	return nil
}
