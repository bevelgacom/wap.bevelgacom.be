package main

import "math/rand"

// NOTE: this is not at all scalable, this should become a database soon

// linkCache is a map of link IDs to their URLs
// this is so we can query cache:linkID and fetch the full URL
// this is done as the Nokia 7110 has a hard link length limit
var linkCache = map[string]string{}
var linkToID = map[string]string{}

func generateID() string {
	// generate a random 8 character string
	randRunes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := ""
	for i := 0; i < 8; i++ {
		// pick random character
		id += string(randRunes[rand.Intn(len(randRunes))])
	}

	return id
}

func StoreLink(link string) string {
	if _, exists := linkToID[link]; exists {
		return linkToID[link]
	}

	id := generateID()
	_, exists := linkCache[id]
	for exists {
		id = generateID()
		_, exists = linkCache[id]
	}
	linkCache[id] = link
	linkToID[link] = id
	return id
}

func GetLink(id string) string {
	return linkCache[id]
}
