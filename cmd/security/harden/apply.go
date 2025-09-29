package harden

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/SlashGordon/nas-manager/internal/constants"
	"github.com/SlashGordon/nas-manager/internal/i18n"
)

func ApplyDSMHardening() error {
	title := i18n.T(i18n.DSMHardeningTitle)
	if _, err := fmt.Fprintf(os.Stdout, "%s\n", title); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat("=", len(title))); err != nil {
		return err
	}

	trusted := false
	modified := false

	if content, err := os.ReadFile(constants.SynoSecurityScanConf); err == nil {
		config := string(content)
		if !strings.Contains(config, "enable_autoblock=yes") {
			choice := constants.ChoiceYes
			if !trusted {
				choice = PromptUser(i18n.T(i18n.DSMAutoBlock))
			}

			switch choice {
			case constants.ChoiceYes:
				if _, err := fmt.Fprintf(os.Stdout, "Auto-applying: %s\n", i18n.T(i18n.DSMAutoBlock)); err != nil {
					return err
				}
				fallthrough
			case constants.ChoiceTrust:
				if choice == constants.ChoiceTrust {
					if _, err := fmt.Fprintf(os.Stdout, "  %s\n", i18n.T(i18n.TrustingRemaining)); err != nil {
						return err
					}
				}

				newConfig := strings.ReplaceAll(config, "enable_autoblock=no", "enable_autoblock=yes")
				if err := os.WriteFile(constants.SynoSecurityScanConf, []byte(newConfig), 0600); err == nil {
					if _, err := fmt.Fprintf(os.Stdout, "%s\n", i18n.T(i18n.DSMAutoBlockEnabled)); err != nil {
						return err
					}
					modified = true
				} else {
					if _, err := fmt.Fprintf(os.Stderr, "%s: %v\n", i18n.T(i18n.DSMAutoBlockFailed), err); err != nil {
						return err
					}
				}
			default:
				if _, err := fmt.Fprintf(os.Stdout, "%s\n", i18n.T(i18n.DSMAutoBlockSkipped)); err != nil {
					return err
				}
			}
		} else {
			if _, err := fmt.Fprintf(os.Stdout, "%s\n", i18n.T(i18n.DSMConfigured)); err != nil {
				return err
			}
		}
	}

	if modified {
		if _, err := fmt.Fprintf(os.Stdout, "\n%s\n", i18n.T(i18n.DSMRestart)); err != nil {
			return err
		}
	}

	return nil
}

func ApplyServiceHardening() error {
	title := i18n.T(i18n.ServiceHardeningTitle)
	if _, err := fmt.Fprintf(os.Stdout, "%s\n", title); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat("=", len(title))); err != nil {
		return err
	}

	services := []string{
		"pkgctl-AudioStation",
		"pkgctl-VideoStation",
		"pkgctl-PhotoStation",
		"pkgctl-SurveillanceStation",
	}

	trusted := false

	for _, service := range services {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "synoservice", "--status", service)
		if output, err := cmd.Output(); err == nil && strings.Contains(string(output), "start") {
			name := strings.TrimPrefix(service, "pkgctl-")
			choice := constants.ChoiceYes
			if !trusted {
				choice = PromptUser(fmt.Sprintf("Disable %s service", name))
			}

			switch choice {
			case constants.ChoiceYes:
				if _, err := fmt.Fprintf(os.Stdout, i18n.T(i18n.ServiceDisabling)+"\n", name); err != nil {
					return err
				}
				fallthrough
			case constants.ChoiceTrust:
				if choice == constants.ChoiceTrust {
					if _, err := fmt.Fprintf(os.Stdout, "  %s\n", i18n.T(i18n.TrustingRemaining)); err != nil {
						return err
					}
				}

				if err := exec.CommandContext(ctx, "synoservice", "--disable", service).Run(); err == nil {
					if _, err := fmt.Fprintf(os.Stdout, i18n.T(i18n.ServiceDisabled)+"\n", name); err != nil {
						return err
					}
				} else {
					if _, err := fmt.Fprintf(os.Stderr, i18n.T(i18n.ServiceFailed)+"\n", name, err); err != nil {
						return err
					}
				}
			default:
				if _, err := fmt.Fprintf(os.Stdout, i18n.T(i18n.ServiceSkipped)+"\n", name); err != nil {
					return err
				}
			}
		} else {
			name := strings.TrimPrefix(service, "pkgctl-")
			if _, err := fmt.Fprintf(os.Stdout, i18n.T(i18n.ServiceAlreadyDisabled)+"\n", name); err != nil {
				return err
			}
		}
	}

	return nil
}
