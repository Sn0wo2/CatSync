package config

func GetDefaultConfig() *Config {
	return &Config{
		Log: Log{
			Level:      "debug",
			Dir:        "./logs",
			FileFormat: "2006-01-02.log",
		},
		Server: Server{
			Address: ":3000",
			Header:  "CatSync",
		},
		Actions: []Action{
			{
				Route:     "/",
				Operation: OperationString,
				ActionOperationString: &StringData{
					Content: "Hello, CatSync!",
				},
			},
		},
	}
}
