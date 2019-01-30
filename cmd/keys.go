package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dmolesUC3/cos/internal/keys"

	"github.com/dmolesUC3/cos/internal/logging"
	"github.com/dmolesUC3/cos/pkg"
)

const (
	usageKeys = "keys <BUCKET-URL>"

	shortDescKeys = "keys: test creating, retrieving, verifying and deleting potentially problematic keys"

	longDescKeys = shortDescKeys + `

		Creates, retrieves, verifies, and deletes a small object for each 
        value in the Big List of Naughty Strings
        (https://github.com/minimaxir/big-list-of-naughty-strings).
	`
	
	exampleKeys = `
		cos keys s3://www.dmoles.net/ --endpoint https://s3.us-west-2.amazonaws.com/
	`
)

type keysFlags struct {
	CosFlags

	From int
	To   int
}

func (f keysFlags) Pretty() string {
	format := `
		log level: %v
		region:   '%v'
		endpoint: '%v'
		from:      %d
        to:        %d
	`
	format = logging.Untabify(format, "  ")

	return fmt.Sprintf(format, f.LogLevel(), f.Region, f.Endpoint, f.From, f.To)
}

func init() {
	flags := keysFlags{}
	cmd := &cobra.Command{
		Use:           usageKeys,
		Short:         shortDescKeys,
		Long:          logging.Untabify(longDescKeys, ""),
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		Example:       logging.Untabify(exampleKeys, "  "),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkKeys(args[0], flags)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		},
	}
	cmdFlags := cmd.Flags()
	flags.AddTo(cmdFlags)
	cmdFlags.IntVarP(&flags.From, "from", "f", 1, "first key to check (1-indexed, inclusive)")
	cmdFlags.IntVarP(&flags.To, "to", "t", -1, "last key to check (1-indexed, inclusive); -1 to check all keys")
	rootCmd.AddCommand(cmd)
}

func checkKeys(bucketStr string, f keysFlags) error {
	logger := f.NewLogger()
	logger.Tracef("flags: %v\n", f)
	logger.Tracef("bucket URL: %v\n", bucketStr)

	source := keys.NaughtyStrings()

	startIndex := f.From - 1
	endIndex := f.To
	if endIndex <= 0 {
		endIndex = source.Count()
	}
	logger.Tracef("startIndex: %d, endIndex: %d\n", startIndex, endIndex)

	k := pkg.NewKeys(f.Endpoint, f.Region, bucketStr, logger)
	failures, err := k.CheckAll(source, startIndex, endIndex)
	if err != nil {
		return err
	}
	failureCount := len(failures)
	if failureCount > 0 {
		totalExpected := endIndex - startIndex
		return fmt.Errorf("%d of %d keys failed", failureCount, totalExpected)
	}
	return nil
}