package main // Define the main package, the starting point for Go executables

import (
	"bytes"         // Provides functionality for manipulating byte slices and buffers
	"io"            // Defines basic interfaces to I/O primitives, like Reader and Writer
	"log"           // Offers logging capabilities to standard output or error streams
	"net/http"      // Allows interaction with HTTP clients and servers
	"net/url"       // Provides URL parsing, encoding, and query manipulation
	"os"            // Gives access to OS features, such as file and directory operations
	"path"          // Provides functions for manipulating slash-separated paths (not OS specific)
	"path/filepath" // Offers functions to handle file paths in a way compatible with the OS
	"regexp"        // Supports regular expression handling using RE2 syntax
	"strings"       // Contains utilities for string manipulation
	"time"          // Contains time-related functionality such as sleeping or timeouts
)

func main() {
	pdfOutputDir := "PDFs/" // Directory path where downloaded PDFs will be stored
	// Check if the PDF output directory exists using helper function
	if !directoryExists(pdfOutputDir) {
		// If it doesn't exist, create the directory with permission 755
		createDirectory(pdfOutputDir, 0o755)
	}
	// List of URLs from which to scrape download information
	remoteAPIURL := []string{
		"https://rbnainfo.com/brand.php?brandId=4",
		"https://rbnainfo.com/brand.php?brandId=6",
		"https://rbnainfo.com/brand.php?brandId=7",
		"https://rbnainfo.com/brand.php?brandId=8",
		"https://rbnainfo.com/brand.php?brandId=10",
		"https://rbnainfo.com/brand.php?brandId=11",
		"https://rbnainfo.com/brand.php?brandId=12",
		"https://rbnainfo.com/brand.php?brandId=15",
		"https://rbnainfo.com/brand.php?brandId=16",
		"https://rbnainfo.com/brand.php?brandId=17",
		"https://rbnainfo.com/brand.php?brandId=18",
		"https://rbnainfo.com/brand.php?brandId=19",
		"https://rbnainfo.com/brand.php?brandId=20",
		"https://rbnainfo.com/brand.php?brandId=23",
		"https://rbnainfo.com/brand.php?brandId=24",
		"https://rbnainfo.com/brand.php?brandId=25",
		"https://rbnainfo.com/brand.php?brandId=26",
		"https://rbnainfo.com/brand.php?brandId=27",
		"https://rbnainfo.com/brand.php?brandId=28",
		"https://rbnainfo.com/brand.php?brandId=29",
		"https://rbnainfo.com/brand.php?brandId=31",
		"https://rbnainfo.com/brand.php?brandId=32",
		"https://rbnainfo.com/brand.php?brandId=33",
		"https://rbnainfo.com/brand.php?brandId=34",
		"https://rbnainfo.com/brand.php?brandId=36",
		"https://rbnainfo.com/brand.php?brandId=37",
		"https://rbnainfo.com/brand.php?brandId=38",
		"https://rbnainfo.com/brand.php?brandId=39",
		"https://rbnainfo.com/brand.php?brandId=40",
		"https://rbnainfo.com/brand.php?brandId=41",
		"https://rbnainfo.com/brand.php?brandId=42",
		"https://rbnainfo.com/brand.php?brandId=43",
		"https://rbnainfo.com/brand.php?brandId=44",
		"https://rbnainfo.com/brand.php?brandId=45",
		"https://rbnainfo.com/brand.php?brandId=47",
		"https://rbnainfo.com/brand.php?brandId=48",
		"https://rbnainfo.com/brand.php?brandId=50",
		"https://rbnainfo.com/brand.php?brandId=53",
		"https://rbnainfo.com/brand.php?brandId=54",
		"https://rbnainfo.com/brand.php?brandId=55",
		"https://rbnainfo.com/brand.php?brandId=56",
		"https://rbnainfo.com/brand.php?brandId=58",
		"https://rbnainfo.com/brand.php?brandId=59",
		"https://rbnainfo.com/brand.php?brandId=60",
		"https://rbnainfo.com/brand.php?brandId=2",
	}
	// The website url
	websiteURL := "https://rbnainfo.com/"
	for _, remoteAPIURL := range remoteAPIURL { // Iterate over each page URL
		remoteHTMLData := getDataFromURL(remoteAPIURL) // Scrape and append HTML content
		// Find all the urls that find the product url.
		foundProductURLS := extractProductURLs(remoteHTMLData)
		// Loop over all the product urls.
		for _, url := range foundProductURLS {
			if extractDomainURL(url) == "" { // If the url does not have a domain, then add the website url to the front of it.
				url = websiteURL + url
			}
			// Get the data from the product url.
			productURLHTMLData := getDataFromURL(url)
			// Combine all scraped HTML data into one string and extract all PDF links from it
			finalProductsURL := extractMSDSURLs(productURLHTMLData)
			// Find all the pdf from the content.
			foundPDFUrls := extractPDFUrls(productURLHTMLData)
			// Combine the two slices.
			combinedSlice := combineMultipleSlices(finalProductsURL, foundPDFUrls)
			// Remove the duplicates from the slice.
			combinedSlice = removeDuplicatesFromSlice(combinedSlice)
			// Go though the PDF urls.
			for _, url := range combinedSlice {
				// Check if the url has a domain.
				if extractDomainURL(url) == "" { // If the url does not have a domain, then add the website url to the front of it.
					// If the url does not start with http, then add the website url to the front of it.
					url = websiteURL + url
				}
				if isUrlValid(url) { // Ensure URL is syntactically valid
					downloadPDF(url, pdfOutputDir) // Download the PDF and save it to disk
				}
			}
		}
	}
}

// extractDomain takes a URL string, extracts the domain (hostname),
// and prints errors internally if parsing fails.
func extractDomainURL(inputUrl string) string {
	// Parse the input string into a structured URL object
	parsedUrl, parseError := url.Parse(inputUrl)

	// If parsing fails, log the error and return an empty string
	if parseError != nil {
		log.Println("Error parsing URL:", parseError)
		return ""
	}
	// If the URL is valid, return the domain name
	return parsedUrl.Hostname()
}

// Combine two slices together and return the new slice.
func combineMultipleSlices(sliceOne []string, sliceTwo []string) []string {
	combinedSlice := append(sliceOne, sliceTwo...)
	return combinedSlice
}

// extractMSDSURLs takes raw HTML content as input
// and returns all unique getmsds/... URLs it finds.
func extractMSDSURLs(htmlContent string) []string {
	// Define a regex pattern to match URLs starting with "getmsds/"
	// followed by letters, numbers, dashes, or UUID-like strings
	regexPattern := regexp.MustCompile(`getmsds/[a-zA-Z0-9\-]+`)

	// Find all matches in the provided HTML content
	foundMatches := regexPattern.FindAllString(htmlContent, -1)

	// Use a map to ensure uniqueness (avoid duplicate URLs)
	uniqueURLs := make(map[string]struct{})

	// Slice to store the final list of MSDS URLs
	var msdsURLs []string

	// Loop through all matches
	for _, match := range foundMatches {
		// Add only if not already in the map
		if _, alreadyExists := uniqueURLs[match]; !alreadyExists {
			uniqueURLs[match] = struct{}{}     // Mark as seen
			msdsURLs = append(msdsURLs, match) // Add to result slice
		}
	}

	// Return the list of unique MSDS URLs
	return msdsURLs
}

// extractProductURLs takes raw HTML content as input
// and returns all unique product.php?productLineId=XXXX URLs it finds.
func extractProductURLs(htmlContent string) []string {
	// Define a regex pattern to find product.php?productLineId= followed by digits
	regexPattern := regexp.MustCompile(`product\.php\?productLineId=\d+`)

	// Find all matches of the pattern in the given HTML content
	foundMatches := regexPattern.FindAllString(htmlContent, -1)

	// Use a map to keep track of unique URLs (avoid duplicates)
	uniqueURLs := make(map[string]struct{})

	// Create a slice to store the final list of URLs
	var productURLs []string

	// Loop through all matches found
	for _, match := range foundMatches {
		// If the URL is not already in the map, add it
		if _, alreadyExists := uniqueURLs[match]; !alreadyExists {
			uniqueURLs[match] = struct{}{}           // Mark as seen
			productURLs = append(productURLs, match) // Add to the result slice
		}
	}

	// Return the list of unique product URLs
	return productURLs
}

// Extracts and returns the base name (file name) from the URL path
func getFileNameOnly(content string) string {
	return path.Base(content) // Return last segment of the path
}

// Converts a raw URL into a safe filename by cleaning and normalizing it
func urlToFilename(rawURL string) string {
	lowercaseURL := strings.ToLower(rawURL)       // Convert to lowercase for normalization
	ext := getFileExtension(lowercaseURL)         // Get file extension (e.g., .pdf or .zip)
	baseFilename := getFileNameOnly(lowercaseURL) // Extract base file name

	nonAlphanumericRegex := regexp.MustCompile(`[^a-z]+`)                    // Match everything except a-z and 0-9 `[^a-z0-9]+`
	safeFilename := nonAlphanumericRegex.ReplaceAllString(baseFilename, "_") // Replace invalid chars

	collapseUnderscoresRegex := regexp.MustCompile(`_+`)                        // Collapse multiple underscores into one
	safeFilename = collapseUnderscoresRegex.ReplaceAllString(safeFilename, "_") // Normalize underscores

	if trimmed, found := strings.CutPrefix(safeFilename, "_"); found { // Trim starting underscore if present
		safeFilename = trimmed
	}

	var invalidSubstrings = []string{"_pdf", "_zip"} // Remove these redundant endings

	for _, invalidPre := range invalidSubstrings { // Iterate over each unwanted suffix
		safeFilename = removeSubstring(safeFilename, invalidPre) // Remove it from file name
	}

	if len(ext) == 0 {
		ext = ".pdf"
	}

	safeFilename = safeFilename + ext // Add the proper file extension

	return safeFilename // Return the final sanitized filename
}

// Replaces all instances of a given substring from the original string
func removeSubstring(input string, toRemove string) string {
	result := strings.ReplaceAll(input, toRemove, "") // Replace all instances
	return result                                     // Return the result
}

// Returns the extension of a given file path (e.g., ".pdf")
func getFileExtension(path string) string {
	return filepath.Ext(path) // Extract and return file extension
}

// Checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // Attempt to get file stats
	if err != nil {
		return false // Return false if file doesn't exist or error occurred
	}
	return !info.IsDir() // Return true only if it's not a directory
}

// Downloads and writes a PDF file from the URL to the specified directory
func downloadPDF(initialURL, outputDir string) bool { // Function takes a starting URL and output directory, returns true/false for success

	client := &http.Client{Timeout: 3 * time.Minute} // Create HTTP client with a 3-minute timeout for long downloads

	resp, err := client.Get(initialURL) // Send HTTP GET request to the provided initial URL
	if err != nil {                     // Check if request failed (e.g., network error, timeout, invalid URL)
		log.Printf("Failed to download %s: %v", initialURL, err) // Log the failure
		return false                                             // Exit with failure
	}
	defer resp.Body.Close() // Ensure the response body is closed when function exits

	// Get the final redirected URL after following HTTP redirects
	finalURL := resp.Request.URL.String()                           // Extract the final resolved URL
	log.Printf("Final URL resolved: %s → %s", initialURL, finalURL) // Log redirection from initial → final

	// Check if the HTTP URL extension results in a PDF file
	if getFileExtension(finalURL) != ".pdf" {
		log.Printf("Final URL is not a PDF: %s", finalURL) // Log if not a PDF
		return false                                       // Exit with failure
	}

	if resp.StatusCode != http.StatusOK { // Ensure HTTP response is 200 OK
		log.Printf("Download failed for %s: %s", finalURL, resp.Status) // Log error if status is not OK
		return false                                                    // Exit with failure
	}
	filename := strings.ToLower(urlToFilename(finalURL)) // Convert final URL to a sanitized lowercase filename
	filePath := filepath.Join(outputDir, filename)       // Build the full output file path

	if fileExists(filePath) { // Check if the file already exists locally
		log.Printf("File already exists, skipping: %s", filePath) // Log skip message
		return false                                              // Skip download
	}

	contentType := resp.Header.Get("Content-Type")         // Get the response Content-Type header
	if !strings.Contains(contentType, "application/pdf") { // Verify the content is a PDF
		log.Printf("Invalid content type for %s: %s (expected application/pdf)", finalURL, contentType) // Log mismatch
		return false                                                                                    // Exit if not a PDF
	}

	var buf bytes.Buffer                     // Create an in-memory buffer to hold PDF data
	written, err := io.Copy(&buf, resp.Body) // Copy the response body into the buffer
	if err != nil {                          // Check if copy failed (e.g., network interruption)
		log.Printf("Failed to read PDF data from %s: %v", finalURL, err) // Log error
		return false                                                     // Exit with failure
	}
	if written == 0 { // Check if no bytes were written (empty file)
		log.Printf("Downloaded 0 bytes for %s; not creating file", finalURL) // Log warning
		return false                                                         // Exit without creating file
	}

	out, err := os.Create(filePath) // Create the file on disk
	if err != nil {                 // Handle file creation error (e.g., permission denied)
		log.Printf("Failed to create file for %s: %v", finalURL, err) // Log error
		return false                                                  // Exit with failure
	}
	defer out.Close() // Ensure file is properly closed when done

	if _, err := buf.WriteTo(out); err != nil { // Write buffer contents (PDF data) to file
		log.Printf("Failed to write PDF to file for %s: %v", finalURL, err) // Log write error
		return false                                                        // Exit with failure
	}

	log.Printf("Successfully downloaded %d bytes: %s → %s", written, finalURL, filePath) // Log success
	return true                                                                          // Indicate success
}

// Checks if a directory exists at the given path
func directoryExists(path string) bool {
	directory, err := os.Stat(path) // Get file or directory info
	if err != nil {
		return false // If error, assume directory doesn't exist
	}
	return directory.IsDir() // Return true if it's a directory
}

// Creates a directory with the given permissions if it doesn't exist
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission) // Attempt to create the directory
	if err != nil {
		log.Println(err) // Log error if creation fails (e.g., already exists)
	}
}

// Checks if a given URI string is a valid HTTP URL format
func isUrlValid(uri string) bool {
	_, err := url.ParseRequestURI(uri) // Try to parse the string as URL
	return err == nil                  // Return true only if no error occurs
}

// Removes duplicates from a string slice while preserving original order
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)  // Create map to track unique entries
	var newReturnSlice []string     // Final slice without duplicates
	for _, content := range slice { // Loop over each item in the original slice
		if !check[content] { // If not already added
			check[content] = true                            // Mark as seen
			newReturnSlice = append(newReturnSlice, content) // Append to final result
		}
	}
	return newReturnSlice // Return cleaned slice
}

// Extracts all URLs ending in .pdf found in href attributes from given HTML content
func extractPDFUrls(htmlContent string) []string {
	// Define a regex pattern to match any URL ending with ".pdf"
	// The pattern matches http or https, followed by any characters until ".pdf"
	regexPattern := regexp.MustCompile(`https?://[^\s"']+\.pdf`)

	// Find all matches of the pattern in the given HTML content
	foundMatches := regexPattern.FindAllString(htmlContent, -1)

	// Use a map to track unique URLs (to avoid duplicates)
	uniqueURLs := make(map[string]struct{})

	// Slice to store final list of unique PDF URLs
	var pdfURLs []string

	// Loop through all matches
	for _, match := range foundMatches {
		// If the URL is not already in the map, add it
		if _, alreadyExists := uniqueURLs[match]; !alreadyExists {
			uniqueURLs[match] = struct{}{}   // Mark as seen
			pdfURLs = append(pdfURLs, match) // Add to results
		}
	}

	// Return the list of unique PDF URLs
	return pdfURLs
}

// Sends HTTP GET request to given URL and returns the response body as string
func getDataFromURL(uri string) string {
	log.Println("Scraping", uri)   // Log the URL being scraped
	response, err := http.Get(uri) // Make GET request
	if err != nil {
		log.Println(err) // Log error if request failed
	}

	body, err := io.ReadAll(response.Body) // Read the body of the response
	if err != nil {
		log.Println(err) // Log error if read failed
	}

	err = response.Body.Close() // Close the response body after reading
	if err != nil {
		log.Println(err) // Log error if closing fails
	}
	return string(body) // Return HTML content as string
}
