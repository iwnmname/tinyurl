package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	serverURL string
	alias     string
	ttlDays   int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tinyurl",
		Short: "TinyURL CLI - сокращайте ссылки из командной строки",
	}

	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "http://localhost:8080", "Адрес сервера TinyURL")

	shortCmd := &cobra.Command{
		Use:   "short [url]",
		Short: "Сократить URL",
		Args:  cobra.ExactArgs(1),
		RunE:  shortURL,
	}
	shortCmd.Flags().StringVarP(&alias, "alias", "a", "", "Пользовательский алиас для ссылки")
	shortCmd.Flags().IntVarP(&ttlDays, "ttl", "t", 0, "Срок жизни ссылки в днях (0 = бессрочно)")

	statsCmd := &cobra.Command{
		Use:   "stats [code]",
		Short: "Получить статистику по коду",
		Args:  cobra.ExactArgs(1),
		RunE:  getStats,
	}

	rootCmd.AddCommand(shortCmd, statsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func shortURL(cmd *cobra.Command, args []string) error {
	url := args[0]
	reqBody, err := json.Marshal(map[string]interface{}{
		"url":      url,
		"alias":    alias,
		"ttl_days": ttlDays,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/shorten", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("сервер вернул ошибку %d: %s", resp.StatusCode, body)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Println("Короткая ссылка:", result["short_url"])
	return nil
}

func getStats(cmd *cobra.Command, args []string) error {
	code := args[0]
	resp, err := http.Get(serverURL + "/stats/" + code)
	if err != nil {
		return fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("сервер вернул ошибку %d: %s", resp.StatusCode, body)
	}

	var stats struct {
		URL       string  `json:"url"`
		CreatedAt string  `json:"created_at"`
		ExpiresAt *string `json:"expires_at,omitempty"`
		HitCount  int     `json:"hit_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return err
	}

	fmt.Println("URL:", stats.URL)
	fmt.Println("Создано:", stats.CreatedAt)
	if stats.ExpiresAt != nil {
		fmt.Println("Истекает:", *stats.ExpiresAt)
	} else {
		fmt.Println("Истекает: никогда")
	}
	fmt.Println("Количество переходов:", stats.HitCount)
	return nil
}
