package main

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/argp"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/token"
	"math/rand"
	"os"
	"time"
)

func main() {
	println("Let's goooo!")

	flags := argp.CreateFlagSet()

	locations := flags.String(argp.FlagBuildArgs{
		Names: []string{"locations", "l"},
		Usage: "The locations file to load the locations data from",
	}, "locations.xml")
	frontendPort := flags.Uint(argp.FlagBuildArgs{
		Names: []string{"frontend-port", "login-port", "lp"},
		Usage: "The port to use for the frontend (login/logout) webserver",
	}, 4443)
	frontendBaseUrlDefaultText := "https://localhost:<frontend-port>/"
	frontendBaseUrl := flags.String(argp.FlagBuildArgs{
		Names:       []string{"frontend-base-url", "base-url"},
		Usage:       "The base url for the frontend server",
		DefaultText: &frontendBaseUrlDefaultText,
	}, "")
	cookieSecretDefaultText := "<random string>"
	cookieSecretArg := flags.String(argp.FlagBuildArgs{
		Names:       []string{"frontend-cookie-secret", "cookie-secret", "cs"},
		Usage:       "The secret used to verify the cookies handed out to the clients",
		DefaultText: &cookieSecretDefaultText,
	}, "")
	backendPort := flags.Uint(argp.FlagBuildArgs{
		Names: []string{"backend-port", "qr-port", "qp"},
		Usage: "The port to use for the backend (QR) webserver",
	}, 443)

	tokenValidTime := flags.Int(argp.FlagBuildArgs{
		Names: []string{"token-valid-time", "valid-time"},
		Usage: "The time that a token is valid for, in seconds",
	}, 120)
	tokenEncryptionSecretDefaultText := "<random string>"
	tokenEncryptionKey := flags.String(argp.FlagBuildArgs{
		Names: []string{"token-encryption-key", "token-encryption-secret", "token-secret"},
		Usage: "The secret that gets used to generate and verify the tokens.\n" +
			"Must be 32 bytes long.",
		DefaultText: &tokenEncryptionSecretDefaultText,
	}, "")

	journalDirectory := flags.String(argp.FlagBuildArgs{
		Names: []string{"journals-directory", "journals", "j"},
		Usage: "The directory to store the journal files in",
	}, "journals")
	journalFilePermissions := flags.Int(argp.FlagBuildArgs{
		Names: []string{"journal-file-permissions"},
		Usage: "Sets the file permission mask for new journal files",
	}, 0777)

	err := flags.ParseFlags(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())

	err = journal.ReadLocations(*locations)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to read locations file: %v", err)
		os.Exit(1)
	}

	if *frontendBaseUrl == "" {
		*frontendBaseUrl = fmt.Sprintf("https://localhost:%v/", *frontendPort)
	}
	logIOUrl = *frontendBaseUrl
	if *cookieSecretArg == "" {
		*cookieSecretArg = randomString(32)
	}
	cookieSecret = *cookieSecretArg

	token.ValidTime = int64(*tokenValidTime)
	if *tokenEncryptionKey == "" {
		*tokenEncryptionKey = randomString(32)
	} else if len(*tokenEncryptionKey) != 32 {
		_, _ = fmt.Fprintf(os.Stderr, "Token encryption key must be 32 characters long")
		os.Exit(1)
	}
	token.EncryptionKey = *tokenEncryptionKey

	dataJournal, err = journal.NewWriter(*journalDirectory)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Couldn't create journal: %v", err)
		os.Exit(1)
	}
	journal.FileCreationPermissions = *journalFilePermissions

	err = RunWebservers(*frontendPort, *backendPort)
	if err != nil {
		fmt.Printf("Couldn't start the Webservers: %#v", err)
	}
}

func randomString(length int) string {
	randBytes := make([]byte, length)
	_, err := rand.Read(randBytes)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to generate a random string: %v", err)
	}
	return string(randBytes)
}
