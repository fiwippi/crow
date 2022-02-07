package api

import (
	"testing"
)

func TestStaticAssetValidation(t *testing.T) {
	_, err := mc.GetStaticAsset("css/yotsubluenew.699.css")
	if err == ErrInvalidAssetFormat {
		t.Errorf("invalid asset format raised for valid asset format: %s\n", "css/yotsubluenew.699.css")
	}

	_, err = mc.GetStaticAsset("image/contest_banners/2fcd223d96df00b4a45d6f79b90035e56f6746cd.jpg")
	if err == ErrInvalidAssetFormat {
		t.Errorf("invalid asset format raised for valid asset format: %s\n", "image/contest_banners/2fcd223d96df00b4a45d6f79b90035e56f6746cd.jpg")
	}

	_, err = mc.GetStaticAsset("js/extension.min.1132.js")
	if err == ErrInvalidAssetFormat {
		t.Errorf("invalid asset format raised for valid asset format: %s\n", "js/extension.min.1132.js")
	}

	_, err = mc.GetStaticAsset("image/favicon-ws.ico")
	if err == ErrInvalidAssetFormat {
		t.Errorf("invalid asset format raised for valid asset format: %s\n", "image/favicon-ws.ico")
	}

	_, err = mc.GetStaticAsset("INVALID")
	if err != ErrInvalidAssetFormat {
		t.Errorf("invalid asset format not raised for invalid asset format: %s\n", "INVALID")
	}
}

func TestSpoilerValidation(t *testing.T) {
	m, err := mc.GetCustomSpoiler("tv", 1)
	if m == nil {
		t.Errorf("failed to load spoiler '1' from 'tv', err: %s\n", err)
	}

	m, err = mc.GetCustomSpoiler("tv", 5)
	if m == nil {
		t.Errorf("failed to load spoiler '5' from 'tv', err: %s\n", err)
	}

	m, err = mc.GetCustomSpoiler("tv", 0)
	if err == nil {
		t.Errorf("error not returned for invalid spoiler '0' from 'tv', err: %s\n", err)
	}

	m, err = mc.GetCustomSpoiler("tv", 0)
	if err == nil {
		t.Errorf("error not returned for invalid spoiler '6' from 'tv', err: %s\n", err)
	}

	m, err = mc.GetCustomSpoiler("g", 1)
	if err == nil {
		t.Errorf("error not returned for invalid spoiler '1' from 'g', err: %s\n", err)
	}
}

func TestMD5Hash(t *testing.T) {
	board := "/po/"
	threadNum := 570368

	thread, _, err := mc.GetThread(board, threadNum)
	if thread == nil {
		t.Errorf("failed to load thread: %s - %d: %s\n", board, threadNum, err)
		return
	}
	post := thread.Posts[0]
	if post == nil {
		t.Error("post in thread is nil")
		return
	}
	m, err := mc.GetFile(post)
	if m == nil {
		t.Errorf("failed to load image from first post in thread: /po/ - 570368, err: %s\n", err)
		return
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
	board := "/po/"
	threadNum := 570368

	thread, _, err := mc.GetThread(board, threadNum)
	if thread == nil {
		t.Errorf("failed to load thread: %s - %d: %s\n", board, threadNum, err)
		return
	}
	post := thread.Posts[0]
	if post == nil {
		t.Error("post in thread is nil")
		return
	}

	id := post.ImageID.String()
	ext := post.Ext
	t.Logf("post Image ID: %s, Ext: %s", id, ext)

	m, err := mc.GetFile(post)
	if m == nil {
		t.Errorf("failed to load image from first post in thread: /po/ - 570368, err: %s\n", err)
	}

	m, err = mc.GetThumbnail(post)
	if m == nil {
		t.Errorf("failed to load image thumbnail from first post in thread: /po/ - 570368, err: %s\n", err)
	}

	m, err = mc.GetFlag("be")
	if m == nil {
		t.Errorf("failed to load flag code 'be', err: %s\n", err)
	}

	m, err = mc.GetTrollFlag("tr")
	if m == nil {
		t.Errorf("failed to load flag code 'tr', err: %s\n", err)
	}

	m, err = mc.GetCustomSpoiler("tv", 1)
	if m == nil {
		t.Errorf("failed to load custom spoiler '1' for board 'tv', err: %s\n", err)
	}

	m, err = mc.GetStaticAsset("css/yotsubluenew.699.css")
	if m == nil {
		t.Errorf("failed to load static asset '%s', err: %s\n", "css/yotsubluenew.699.css", err)
	}
}
