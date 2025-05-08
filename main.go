package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	ConfigFile string
	EnvFile    string
	HyprFile   string
	NoDaemon   bool
	NoExport   bool
	Verbose    bool
	Debug      bool
}

type TomlMap map[string]interface{}

type EnvVars []string
type HyprVars []string

var (
	config Config
)

func main() {

	flag.StringVar(&config.ConfigFile, "input", filepath.Join(xdg.ConfigHome, "hyde", "config.toml"),
		"The input TOML file to parse. Default is $XDG_CONFIG_HOME/hyde/config.toml")
	flag.StringVar(&config.EnvFile, "env", filepath.Join(xdg.StateHome, "hyde", "config"),
		"The output environment file. Default is $XDG_STATE_HOME/hyde/config")
	flag.StringVar(&config.HyprFile, "hypr", filepath.Join(xdg.StateHome, "hyde", "hyprland.conf"),
		"The output Hyprland file. Default is $XDG_STATE_HOME/hyde/hyprland.conf")
	flag.BoolVar(&config.NoDaemon, "no-daemon", false, "Run in one-off mode without watching for changes (daemon mode is default)")
	flag.BoolVar(&config.NoExport, "no-export", false, "Disable exporting the parsed data (export is default)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug mode with detailed logging")
	flag.Parse()

	if config.Verbose || config.Debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	} else {
		log.SetFlags(0)
	}
	log.SetPrefix("hyde-config: ")

	logInfo("Using config file: %s", config.ConfigFile)
	logInfo("Using env output file: %s", config.EnvFile)
	logInfo("Using hypr output file: %s", config.HyprFile)
	logInfo("Export mode: %v", !config.NoExport)
	logInfo("Daemon mode: %v", !config.NoDaemon)
	logInfo("Debug mode: %v", config.Debug)

	ensureDirExists(filepath.Dir(config.EnvFile))
	ensureDirExists(filepath.Dir(config.HyprFile))

	parseConfigFiles(config.ConfigFile, config.EnvFile, config.HyprFile, !config.NoExport)

	if !config.NoDaemon {
		logInfo("Starting daemon mode, watching %s for changes", config.ConfigFile)
		watchFile(config.ConfigFile, config.EnvFile, config.HyprFile, !config.NoExport)
	} else {
		logInfo("Running in one-off mode (no watching for changes)")
	}
}

func ensureDirExists(dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
}

func parseConfigFiles(tomlFile, envFile, hyprFile string, exportMode bool) bool {

	fileInfo, err := os.Stat(tomlFile)
	if err != nil {
		logError("Failed to stat config file: %v", err)
		return false
	}

	if fileInfo.Size() == 0 {
		logError("Config file is empty, skipping parse")
		return false
	}

	tomlContent, err := loadTomlFile(tomlFile)
	if err != nil {
		logError("Failed to load TOML file: %v", err)
		return false
	}
	if len(tomlContent) == 0 {
		logError("TOML content is empty, skipping parse")
		return false
	}

	logDebug("TOML content loaded successfully, size of map: %d", len(tomlContent))

	var wg sync.WaitGroup
	wg.Add(2)

	var success1, success2 bool

	go func() {
		defer wg.Done()
		success1 = parseTomlToEnvWithContent(tomlContent, envFile, exportMode)
	}()

	go func() {
		defer wg.Done()
		success2 = parseTomlToHyprWithContent(tomlContent, hyprFile)
	}()

	wg.Wait()

	return success1 && success2
}

func logInfo(format string, v ...interface{}) {
	if config.Verbose || config.Debug {
		log.Printf(format, v...)
	}
}

func logDebug(format string, v ...interface{}) {
	if config.Debug {
		log.Printf("DEBUG: "+format, v...)
	}
}

func logError(format string, v ...interface{}) {
	log.Printf("ERROR: "+format, v...)
}

func loadTomlFile(tomlFile string) (TomlMap, error) {

	_, err := os.Stat(tomlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to access TOML file: %w", err)
	}

	data, err := os.ReadFile(tomlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read TOML file: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("TOML file is empty")
	}

	logDebug("Read %d bytes from TOML file", len(data))

	var tomlContent TomlMap
	if err := toml.Unmarshal(data, &tomlContent); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	return tomlContent, nil
}

func parseTomlToEnvWithContent(tomlContent TomlMap, envFile string, exportMode bool) bool {
	ignoredKeys := map[string]bool{
		"$schema":        true,
		"$SCHEMA":        true,
		"hyprland":       true,
		"hyprland-ipc":   true,
		"hyprland-env":   true,
		"hyprland-start": true,
	}

	if len(tomlContent) == 0 {
		logError("Cannot parse empty TOML content")
		return false
	}

	envVars := make(EnvVars, 0, len(tomlContent)*2)
	flattenDict(tomlContent, "", ignoredKeys, &envVars, exportMode)

	if len(envVars) == 0 {
		logError("No environment variables generated, skipping file write")
		return false
	}

	logDebug("Generated %d environment variable lines", len(envVars))

	if envFile != "" {

		tempFile := envFile + ".tmp"
		if err := writeLinesToFile(tempFile, envVars); err != nil {
			logError("Failed to write environment variables to temp file: %v", err)
			return false
		}

		if err := os.Rename(tempFile, envFile); err != nil {
			logError("Failed to replace environment file: %v", err)

			os.Remove(tempFile)
			return false
		}

		logInfo("Environment variables have been written to %s", envFile)
		return true
	} else {
		for _, line := range envVars {
			logInfo("%s", line)
		}
		return true
	}
}

func flattenDict(data TomlMap, parentKey string, ignoredKeys map[string]bool, result *EnvVars, exportMode bool) {
	for k, v := range data {

		if ignoredKeys[k] || (parentKey != "" && strings.HasPrefix(parentKey, "hyprland")) {
			logDebug("Skipping ignored key: %s", k)
			continue
		}

		if strings.HasPrefix(k, "$") {
			continue
		}

		var newKey string
		if parentKey != "" {
			newKey = fmt.Sprintf("%s_%s", parentKey, strings.ToUpper(k))
		} else {
			newKey = strings.ToUpper(k)
		}

		switch val := v.(type) {
		case map[string]interface{}:

			flattenDict(TomlMap(val), newKey, ignoredKeys, result, exportMode)
		case []interface{}:

			arrayItems := make([]string, 0, len(val))
			for _, item := range val {
				arrayItems = append(arrayItems, fmt.Sprintf("\"%v\"", item))
			}
			value := fmt.Sprintf("(%s)", strings.Join(arrayItems, " "))
			if exportMode {
				*result = append(*result, fmt.Sprintf("export %s=%s", newKey, value))
			} else {
				*result = append(*result, fmt.Sprintf("%s=%s", newKey, value))
			}
		case bool:

			value := strconv.FormatBool(val)
			if exportMode {
				*result = append(*result, fmt.Sprintf("export %s=%s", newKey, value))
			} else {
				*result = append(*result, fmt.Sprintf("%s=%s", newKey, value))
			}
		case int64, float64:

			if exportMode {
				*result = append(*result, fmt.Sprintf("export %s=%v", newKey, val))
			} else {
				*result = append(*result, fmt.Sprintf("%s=%v", newKey, val))
			}
		default:

			if exportMode {
				*result = append(*result, fmt.Sprintf("export %s=\"%v\"", newKey, val))
			} else {
				*result = append(*result, fmt.Sprintf("%s=\"%v\"", newKey, val))
			}
		}
	}
}

func parseTomlToHyprWithContent(tomlContent TomlMap, hyprFile string) bool {
	if len(tomlContent) == 0 {
		logError("Cannot parse empty TOML content for Hyprland")
		return false
	}

	hyprVars := make(HyprVars, 0, 32)
	flattenHyprDict(tomlContent, "", &hyprVars)

	if len(hyprVars) == 0 {
		logError("No Hyprland variables generated, skipping file write")
		return false
	}

	logDebug("Generated %d Hyprland variable lines", len(hyprVars))

	if hyprFile != "" {

		tempFile := hyprFile + ".tmp"
		if err := writeLinesToFile(tempFile, hyprVars); err != nil {
			logError("Failed to write Hyprland variables to temp file: %v", err)
			return false
		}

		if err := os.Rename(tempFile, hyprFile); err != nil {
			logError("Failed to replace Hyprland file: %v", err)

			os.Remove(tempFile)
			return false
		}

		logInfo("Hyprland variables have been written to %s", hyprFile)
		return true
	} else {
		logInfo("No hypr file specified.")
		for _, line := range hyprVars {
			logInfo("%s", line)
		}
		return true
	}
}

func flattenHyprDict(data TomlMap, parentKey string, result *HyprVars) {
	for k, v := range data {

		isHyprlandSection := strings.HasPrefix(k, "hyprland") || strings.HasPrefix(parentKey, "hyprland")

		if isHyprlandSection {
			logDebug("Found hyprland key: %s", k)

			newKey := k
			if strings.HasPrefix(newKey, "hyprland_") {
				newKey = strings.Replace(newKey, "hyprland_", "", 1)
			}

			if parentKey != "" && !strings.HasPrefix(parentKey, "hyprland") {
				newKey = fmt.Sprintf("%s_%s", parentKey, newKey)
			} else if strings.HasPrefix(parentKey, "hyprland") {
				if len(parentKey) > 9 {
					newKey = fmt.Sprintf("$%s.%s", parentKey[9:], strings.ToUpper(newKey))
				} else {
					newKey = fmt.Sprintf("$%s", strings.ToUpper(newKey))
				}
			}

			switch val := v.(type) {
			case map[string]interface{}:
				flattenHyprDict(TomlMap(val), newKey, result)
			case []interface{}:
				arrayItems := make([]string, 0, len(val))
				for _, item := range val {
					arrayItems = append(arrayItems, fmt.Sprintf("%v", item))
				}
				value := strings.Join(arrayItems, ", ")
				*result = append(*result, fmt.Sprintf("%s=%s", newKey, value))
			case bool:
				*result = append(*result, fmt.Sprintf("%s=%t", newKey, val))
			case int64, float64:
				*result = append(*result, fmt.Sprintf("%s=%v", newKey, val))
			default:
				*result = append(*result, fmt.Sprintf("%s=%v", newKey, val))
			}
		} else {
			logDebug("Skipping key: %s", k)
		}
	}
}

func writeLinesToFile(filename string, lines []string) error {
	if len(lines) == 0 {
		return fmt.Errorf("no lines to write")
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush file buffer: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file to disk: %w", err)
	}

	return nil
}

var (
	watcherMutex sync.Mutex
	lastMod      time.Time
)

func watchFile(tomlFile, envFile, hyprFile string, exportMode bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logError("Failed to create file watcher: %v", err)
		return
	}
	defer watcher.Close()

	configDir := filepath.Dir(tomlFile)
	err = watcher.Add(configDir)
	if err != nil {
		logError("Failed to watch directory %s: %v", configDir, err)
		return
	}

	logInfo("Watching directory %s for changes to %s", configDir, filepath.Base(tomlFile))

	watcherMutex.Lock()
	lastMod = time.Now()
	watcherMutex.Unlock()

	debounceInterval := 300 * time.Millisecond
	configFileName := filepath.Base(tomlFile)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if filepath.Base(event.Name) != configFileName {
				continue
			}

			logDebug("Received event %s for file %s", event.Op, event.Name)

			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {

				info, err := os.Stat(tomlFile)
				if err != nil {
					logError("Failed to stat file: %v", err)
					continue
				}

				watcherMutex.Lock()
				shouldProcess := time.Since(lastMod) > debounceInterval
				if shouldProcess {
					lastMod = info.ModTime()
				}
				watcherMutex.Unlock()

				if shouldProcess {
					logInfo("Config file changed (size: %d bytes), reprocessing", info.Size())

					time.Sleep(50 * time.Millisecond)

					parseConfigFiles(tomlFile, envFile, hyprFile, exportMode)
				} else {
					logDebug("Skipping event, within debounce interval")
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logError("Watcher error: %v", err)
		}
	}
}
