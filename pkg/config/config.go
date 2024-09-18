package config

import "github.com/spf13/viper"

type Config struct {
	Port          string `mapstructure:"PORT"`
	DBUrl         string `mapstructure:"DB_URL"`
	CommentSvcUrl string `mapstructure:"COMMENT_SVC_URL"`
	GoNewsSvcUrl  string `mapstructure:"GONEWS_SVC_URL"`
}

func LoadConfig() (Config, error) {
	var c Config
	viper.AddConfigPath("./pkg/config/envs")
	viper.AddConfigPath("/GoNew-service") //Для docker

	viper.SetConfigName("prod")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()

	if err != nil {
		viper.SetConfigName("dev")
		viper.SetConfigType("env")
		viper.AutomaticEnv()

		err = viper.ReadInConfig()

		if err != nil {
			return Config{}, err
		}
		err = viper.Unmarshal(&c)

		return c, nil
	}

	err = viper.Unmarshal(&c)

	return c, nil
}
