package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

func init() {
	searchCmd.PersistentFlags().Int("offset", 0, "Offset of the search")
	searchCmd.PersistentFlags().Int("limit", 10, "Limit of the search")
	searchCmd.PersistentFlags().String("order", "desc", "Sort order of the search")
	searchCmd.PersistentFlags().String("order-by", "last_version_date", "Order field of the search")
	searchCmd.PersistentFlags().String("format", "list", "Order field of the search")

	_ = viper.BindPFlag("offset", searchCmd.PersistentFlags().Lookup("offset"))
	_ = viper.BindPFlag("limit", searchCmd.PersistentFlags().Lookup("limit"))
	_ = viper.BindPFlag("order", searchCmd.PersistentFlags().Lookup("order"))
	_ = viper.BindPFlag("order-by", searchCmd.PersistentFlags().Lookup("order-by"))
	_ = viper.BindPFlag("format", searchCmd.PersistentFlags().Lookup("format"))
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search mods",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ficsit.InitAPI()

		search := ""
		if len(args) > 0 {
			search = args[0]
		}

		response, err := ficsit.Mods(cmd.Context(), client, ficsit.ModFilter{
			Limit:    viper.GetInt("limit"),
			Offset:   viper.GetInt("offset"),
			Order:    ficsit.Order(viper.GetString("order")),
			Order_by: ficsit.ModFields(viper.GetString("order-by")),
			Search:   search,
		})
		if err != nil {
			return err
		}

		modList := response.Mods.Mods

		switch viper.GetString("format") {
		default:
			for _, mod := range modList {
				println(fmt.Sprintf("%s (%s)", mod.Name, mod.Mod_reference))
			}
		case "json":
			result, err := json.MarshalIndent(modList, "", "  ")
			if err != nil {
				return errors.Wrap(err, "failed converting mods to json")
			}
			println(string(result))
		}

		return nil
	},
}
