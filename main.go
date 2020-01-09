package main

import (
	"flag"
	"fmt"
	"os"

	prettytable "github.com/jedib0t/go-pretty/table"
	prettytext "github.com/jedib0t/go-pretty/text"
)

const (
	fatCaloriesPerGram     = 9
	proteinCaloriesPerGram = 4
	carbCaloriesPerGram    = 4
)

type config struct {
	PercentCarbs                int
	PercentProtein              int
	PercentFat                  int
	HardCarbLimit               int
	PercentDeficit              int
	Calories                    int
	TotalDailyEnergyExepnditure int
}

type macros struct {
	CarbCalories    int
	CarbGrams       int
	ProteinCalories int
	ProteinGrams    int
	FatCalories     int
	FatGrams        int
}

func parse() *config {
	var cfg config
	flags := flag.NewFlagSet("macrocalc", flag.ExitOnError)
	flags.IntVar(&cfg.PercentCarbs, "pct-carbs", 10, "percentage of macro-nutrients from carbs")
	flags.IntVar(&cfg.PercentProtein, "pct-protein", 30, "percentage of macro-nutrients from protein")
	flags.IntVar(&cfg.PercentFat, "pct-fat", 60, "percentage of macro-nutrients from protein")
	flags.IntVar(&cfg.HardCarbLimit, "carb-limit", -1, "hard limit on number of net carbs per day")
	flags.IntVar(&cfg.PercentDeficit, "pct-deficit", 20, "percentage deficit per day")
	flags.IntVar(&cfg.Calories, "calories", -1, "total calories per day")
	flags.IntVar(&cfg.TotalDailyEnergyExepnditure, "tdee", -1, "total daily energy expenditure")
	flags.Parse(os.Args[1:])
	return &cfg
}

func calculateMacros(cfg *config) macros {
	results := macros{}
	results.CarbGrams = (cfg.PercentCarbs * cfg.Calories) / (100 * carbCaloriesPerGram)
	if cfg.HardCarbLimit > 0 && results.CarbGrams > cfg.HardCarbLimit {
		results.CarbGrams = cfg.HardCarbLimit
		newCarbPct := ((results.CarbGrams * carbCaloriesPerGram * 100) / cfg.Calories)
		diffCarbPct := cfg.PercentCarbs - newCarbPct
		cfg.PercentCarbs = newCarbPct
		cfg.PercentProtein += diffCarbPct
	}
	results.CarbCalories = results.CarbGrams * carbCaloriesPerGram
	results.ProteinGrams = (cfg.PercentProtein * cfg.Calories) / (100 * proteinCaloriesPerGram)
	results.ProteinCalories = results.ProteinGrams * proteinCaloriesPerGram
	results.FatCalories = cfg.Calories - results.CarbCalories - results.ProteinCalories
	results.FatGrams = results.FatCalories / fatCaloriesPerGram
	return results
}

func main() {
	cfg := parse()
	prettytext.DisableColors()
	writer := prettytable.NewWriter()
	writer.AppendHeader(prettytable.Row{"Macro", "Grams", "Percentage", "Calories"})
	writer.SetColumnConfigs([]prettytable.ColumnConfig{
		{Name: "Percentage", Transformer: prettytext.NewNumberTransformer("%4.1f")},
	})
	writer.SetStyle(prettytable.StyleLight)

	if cfg.Calories == -1 && cfg.TotalDailyEnergyExepnditure == -1 {
		fmt.Println("error: must specify either daily calories or total daily energy expenditure")
		os.Exit(1)
	}
	if cfg.TotalDailyEnergyExepnditure > 0 {
		cfg.Calories = cfg.TotalDailyEnergyExepnditure * int(100-cfg.PercentDeficit) / 100
		fmt.Printf("%2d%% of %d is %d\n", cfg.PercentDeficit, cfg.TotalDailyEnergyExepnditure, cfg.Calories)
	}

	results := calculateMacros(cfg)

	writer.AppendRows([]prettytable.Row{
		{"Fat", results.FatGrams, float64(results.FatCalories*100) / float64(cfg.Calories), results.FatCalories},
		{"Protein", results.ProteinGrams, float64(results.ProteinCalories*100) / float64(cfg.Calories), results.ProteinCalories},
		{"Carbs", results.CarbGrams, float64(results.CarbCalories*100) / float64(cfg.Calories), results.CarbCalories},
	})
	writer.AppendFooter(prettytable.Row{"Total", "-", cfg.PercentCarbs + cfg.PercentProtein + cfg.PercentFat, cfg.Calories})
	fmt.Println(writer.Render())
}
