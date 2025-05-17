package release

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/blake2s"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/release/distsign"
)

var cmdSignPackage = &cobra.Command{
	Use:   "sign-package",
	Short: "sign a package with a signing key",
	RunE:  cmdSignPackageFunc,
}

var (
	flagSignPackagePackagePath  string
	flagSignPackageSignPrivPath string
	flagSignPackageSigPath      string
)

func init() {
	cmdSignPackage.PersistentFlags().StringVar(&flagSignPackagePackagePath, "package-path", "", "path of package")
	cmdSignPackage.PersistentFlags().StringVar(&flagSignPackageSignPrivPath, "sign-priv-path", "", "path of signing private key")
	cmdSignPackage.PersistentFlags().StringVar(&flagSignPackageSigPath, "sig-path", "", "output path of signature path")
}

func cmdSignPackageFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting sign-package command")

	signPrivRaw, err := os.ReadFile(flagSignPackageSignPrivPath)
	if err != nil {
		return err
	}
	signPrivKey, err := distsign.ParseSigningKey(signPrivRaw)
	if err != nil {
		return err
	}

	pkgData, err := os.ReadFile(flagSignPackagePackagePath)
	if err != nil {
		return err
	}

	hash := blake2s.Sum256(pkgData)
	sig, err := signPrivKey.SignPackageHash(hash[:], int64(len(pkgData)))
	if err != nil {
		return err
	}

	if err := os.WriteFile(flagSignPackageSigPath, sig, 0400); err != nil {
		return err
	}

	return nil
}
