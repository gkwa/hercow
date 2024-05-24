package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gkwa/hercow/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/taylormonacelli/goldbug"
)

var (
	cfgFile   string
	verbose   bool
	logFormat string
)

var rootCmd = &cobra.Command{
	Use:   "hercow",
	Short: "Recursively replace strings in files and filenames within a Git-controlled directory",
	Long: `Hercow is a command-line tool that recursively searches for a specified string within a
Git-controlled directory and replaces it with a new string in both file contents and filenames.
It provides options to control the maximum number of files processed and enables logging for
debugging purposes.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		maxFiles, _ := cmd.Flags().GetInt("maxfiles")
		replace, _ := cmd.Flags().GetString("replace")
		skipDirs, _ := cmd.Flags().GetStringSlice("skip-dirs")

		if len(args) == 0 {
			fmt.Println("Error: directory path is required")
			os.Exit(1)
		}

		core.Main(args[0], maxFiles, replace, skipDirs)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	var err error

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hercow.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose mode")
	err = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		slog.Error("error binding verbose flag", "error", err)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "json or text (default is text)")
	err = viper.BindPFlag("log-format", rootCmd.PersistentFlags().Lookup("log-format"))
	if err != nil {
		slog.Error("error binding log-format flag", "error", err)
		os.Exit(1)
	}

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	err = viper.BindPFlag("toggle", rootCmd.Flags().Lookup("toggle"))
	if err != nil {
		slog.Error("error binding toggle flag", "error", err)
		os.Exit(1)
	}

	rootCmd.Flags().IntP("maxfiles", "m", 100, "maximum number of files allowed")
	rootCmd.Flags().StringP("replace", "r", "", "string replacement in the format 'string1=string2'")
	rootCmd.Flags().StringSliceP("skip-dirs", "s", []string{".git"}, "directories to skip")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".hercow")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	logFormat = viper.GetString("log-format")
	verbose = viper.GetBool("verbose")

	slog.Debug("using config file", "path", viper.ConfigFileUsed())
	slog.Debug("log-format", "value", logFormat)
	slog.Debug("log-format", "value", viper.GetString("log-format"))

	setupLogging()
}

func setupLogging() {
	if verbose || logFormat != "" {
		if logFormat == "json" {
			goldbug.SetDefaultLoggerJson(slog.LevelDebug)
		} else {
			goldbug.SetDefaultLoggerText(slog.LevelDebug)
		}

		slog.Debug("setup", "verbose", verbose)
	}
}
