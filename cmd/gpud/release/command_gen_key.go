package release

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/release/distsign"
)

var cmdGenKey = &cobra.Command{
	Use:   "gen-key",
	Short: "generate root or signing key pair",
	RunE:  cmdGenKeyFunc,
}

var (
	flagGenKeyRoot     bool
	flagGenKeySigning  bool
	flagGenKeyPrivPath string
	flagGenKeyPubPath  string
)

func init() {
	cmdGenKey.PersistentFlags().BoolVar(&flagGenKeyRoot, "root", false, "generate root key")
	cmdGenKey.PersistentFlags().BoolVar(&flagGenKeySigning, "signing", false, "generate signing key")
	cmdGenKey.PersistentFlags().StringVar(&flagGenKeyPrivPath, "priv-path", "", "path of the private key")
	cmdGenKey.PersistentFlags().StringVar(&flagGenKeyPubPath, "pub-path", "", "path of the public key")
}

func cmdGenKeyFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting gen-key command")

	var pub, priv []byte
	switch {
	case flagGenKeyRoot && flagGenKeySigning:
		return errors.New("only one of --root or --signing can be set")
	case !flagGenKeyRoot && !flagGenKeySigning:
		return errors.New("set either --root or --signing")
	case flagGenKeyRoot:
		priv, pub, err = distsign.GenerateRootKey()
	case flagGenKeySigning:
		priv, pub, err = distsign.GenerateSigningKey()
	}
	if err != nil {
		fmt.Printf("failed to generate key pair: %v\n", err)
		return err
	}

	if err := os.WriteFile(flagGenKeyPrivPath, priv, 0400); err != nil {
		return fmt.Errorf("failed writing private key: %w", err)
	}
	fmt.Println("wrote private key to", flagGenKeyPrivPath)

	if err := os.WriteFile(flagGenKeyPubPath, pub, 0400); err != nil {
		return fmt.Errorf("failed writing public key: %w", err)
	}
	fmt.Println("wrote public key to", flagGenKeyPubPath)

	return nil
}
