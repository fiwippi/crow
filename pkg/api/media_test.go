package api

import (
	"testing"
)

func TestStaticAssetValidation(t *testing.T) {
	_, err := mc.GetStaticAsset("css/yotsubluenew.699.css", HTTP)
	if err == ErrInvalidAssetFormat {
		t.Errorf("Invalid asset format raised for valid asset format: %s\n", "css/yotsubluenew.699.css")
	}

	_, err = mc.GetStaticAsset("image/contest_banners/2fcd223d96df00b4a45d6f79b90035e56f6746cd.jpg", HTTP)
	if err == ErrInvalidAssetFormat {
		t.Errorf("Invalid asset format raised for valid asset format: %s\n", "image/contest_banners/2fcd223d96df00b4a45d6f79b90035e56f6746cd.jpg")
	}

	_, err = mc.GetStaticAsset("js/extension.min.1132.js", HTTP)
	if err == ErrInvalidAssetFormat {
		t.Errorf("Invalid asset format raised for valid asset format: %s\n", "js/extension.min.1132.js")
	}

	_, err = mc.GetStaticAsset("image/favicon-ws.ico", HTTP)
	if err == ErrInvalidAssetFormat {
		t.Errorf("Invalid asset format raised for valid asset format: %s\n", "image/favicon-ws.ico")
	}

	_, err = mc.GetStaticAsset("INVALID", HTTP)
	if err != ErrInvalidAssetFormat {
		t.Errorf("Invalid asset format not raised for invalid asset format: %s\n", "INVALID")
	}
}

func TestSpoilerValidation(t *testing.T) {
	m, err := mc.GetCustomSpoiler(1, "tv", HTTP)
	if m == nil {
		t.Errorf("Failed to load spoiler '1' from 'tv', err: %s\n", err)
	}

	m, err = mc.GetCustomSpoiler(5, "tv", HTTP)
	if m == nil {
		t.Errorf("Failed to load spoiler '5' from 'tv', err: %s\n", err)
	}

	//
	m, err = mc.GetCustomSpoiler(0, "tv", HTTP)
	if err == nil {
		t.Errorf("Error not returned for invalid spoiler '0' from 'tv', err: %s\n", err)
	}

	m, err = mc.GetCustomSpoiler(6, "tv", HTTP)
	if err == nil {
		t.Errorf("Error not returned for invalid spoiler '6' from 'tv', err: %s\n", err)
	}

	m, err = mc.GetCustomSpoiler(1, "g", HTTP)
	if err == nil {
		t.Errorf("Error not returned for invalid spoiler '1' from 'g', err: %s\n", err)
	}
}

func TestMD5Hash(t *testing.T) {
	mc.ResetLastAccessed()

	board := "/po/"
	threadNum := "570368"
	thread, _, _ := mc.GetThread(threadNum, board, false, HTTP)
	if thread == nil {
		t.Errorf("Failed to load thread: %s - %s\n", board, threadNum)
	}

	post := thread.Posts[0]
	if post == nil {
		t.Error("Post in thread is nil")
	}

	m, err := mc.GetFileFromPost(post, HTTP)
	if m == nil {
		t.Errorf("Failed to load image from first post in thread: /po/ - 570368, err: %s\n", err)
	}

	// Ensure the MD5 hash is correct
	if m.MD5 != "uZUeZeB14FVR+Mc2ScHvVA==" {
		t.Errorf("MD5 not hashed correctly")
	}

	// Check they VerifyMD5 can recognise this hash is correct
	if !VerifyMD5(post, m) {
		t.Errorf("MD5 hash not recognised as correct")
	}
}

func TestMediaRequests(t *testing.T) {
	mc.ResetLastAccessed()

	board := "/po/"
	threadNum := "570368"
	thread, _, _ := mc.GetThread(threadNum, board, false, HTTP)
	if thread == nil {
		t.Errorf("Failed to load thread: %s - %s\n", board, threadNum)
	}

	post := thread.Posts[0]
	if post == nil {
		t.Error("Post in thread is nil")
	}
	id := post.ImageID.String()
	ext := post.Ext
	t.Logf("Post Image ID: %s, Ext: %s", id, ext)

	m, err := mc.GetFileFromID(id, ext, board, HTTP)
	if m == nil {
		t.Errorf("Failed to load image from first post in thread: /po/ - 570368, err: %s\n", err)
	}

	m, err = mc.GetFileFromPost(post, HTTP)
	if m == nil {
		t.Errorf("Failed to load image from first post in thread: /po/ - 570368, err: %s\n", err)
	}

	m, err = mc.GetThumbnailFromID(id, board, HTTP)
	if m == nil {
		t.Errorf("Failed to load image thumbnail from first post in thread: /po/ - 570368, err: %s\n", err)
	}

	m, err = mc.GetThumbnailFromPost(post, HTTP)
	if m == nil {
		t.Errorf("Failed to load image thumbnail from first post in thread: /po/ - 570368, err: %s\n", err)
	}

	m, err = mc.GetFlagFromCode("be", HTTP)
	if m == nil {
		t.Errorf("Failed to load flag code 'be', err: %s\n", err)
	}

	m, err = mc.GetTrollFlagFromCode("tr", HTTP)
	if m == nil {
		t.Errorf("Failed to load flag code 'tr', err: %s\n", err)
	}

	m, err = mc.GetCustomSpoiler(1, "tv", HTTP)
	if m == nil {
		t.Errorf("Failed to load custom spoiler '1' for board 'tv', err: %s\n", err)
	}

	m, err = mc.GetStaticAsset("css/yotsubluenew.699.css", HTTP)
	if m == nil {
		t.Errorf("Failed to load static asset '%s', err: %s\n", "css/yotsubluenew.699.css", err)
	}
}
