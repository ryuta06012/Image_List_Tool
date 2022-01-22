/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   main.go                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: hryuuta <hryuuta@student.42tokyo.jp>       +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2022/01/21 11:10:04 by hryuuta           #+#    #+#             */
/*   Updated: 2022/01/22 14:51:36 by hryuuta          ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Allowed POST method only", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fileSrc, _, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer fileSrc.Close()
	fileDest, err := os.Create("original.jpg")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer fileDest.Close()

	io.Copy(fileDest, fileSrc)
	log.Println("Original image was saved.")
	http.Redirect(w, r, "/show", http.StatusFound)
}

func ShowHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("original.jpg")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var decodeImageNames []image.Image
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	decodeImageNames = append(decodeImageNames, img)
	writeImageWithTemplate(w, "show", decodeImageNames)
}

func handleClockTpl(w http.ResponseWriter, r *http.Request) {
	// テンプレートをパース
	dir, err := os.Open("dir/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dir.Close()
	allImageNames, err := dir.Readdirnames(-1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	var decodeAllImages []image.Image
	for _, ImageName := range allImageNames {
		fileName, err := os.Open("dir/" + ImageName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer fileName.Close()
		decodeImage, _, err := image.Decode(fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		decodeAllImages = append(decodeAllImages, decodeImage)
	}
	writeImageWithTemplate(w, "index", decodeAllImages)
}

func writeImageWithTemplate(w http.ResponseWriter, tmpl string, decodeAllImages []image.Image) {
	var encordImages []string
	for _, decodeImage := range decodeAllImages {
		buffer := new(bytes.Buffer)
		if err := jpeg.Encode(buffer, decodeImage, nil); err != nil {
			log.Fatalln("Unable to encode image.")
		}
		str := base64.StdEncoding.EncodeToString(buffer.Bytes())
		encordImages = append(encordImages, str)
	}
	data := map[string]interface{}{"Images": encordImages}
	renderTemplate(w, tmpl, data)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	var templates = template.Must(template.ParseFiles("templates/" + tmpl + ".html"))
	if err := templates.ExecuteTemplate(w, tmpl+".html", data); err != nil {
		log.Fatalln("Unable to execute template.")
	}
}

func main() {
	// /now にアクセスした際に処理するハンドラーを登録
	http.HandleFunc("/now", handleClockTpl)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/show", ShowHandler)

	// サーバーをポート8080で起動
	log.Fatal(http.ListenAndServe(":8080", nil))
}
