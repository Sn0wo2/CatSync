//go:build catsync_all || feature_dotenv

package main

import "github.com/joho/godotenv"

func init() { _ = godotenv.Load() }
