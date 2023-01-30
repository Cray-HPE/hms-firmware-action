/*
 * MIT License
 *
 * (C) Copyright [2020-2023] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package domain

import (
	"errors"
	"net/http"

	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func GetStoredImages() (images []storage.Image, err error) {
	images, err = (*GLOB.DSP).GetImages()
	return
}

func GetStoredImage(imageID uuid.UUID) (image storage.Image, err error) {
	// Do not look for a Nil uuid in the database, just return error
	if imageID == uuid.Nil {
		err = errors.New("Null image id")
		return
	}
	image, err = (*GLOB.DSP).GetImage(imageID)
	return
}

func StoreImage(image storage.Image) (err error) {
	err = (*GLOB.DSP).StoreImage(image)
	return
}

func DeleteStoredImage(imageID uuid.UUID) (err error) {
	err = (*GLOB.DSP).DeleteImage(imageID)
	return
}

// CreateImage - will create an image
func CreateImage(i presentation.RawImage) (pb model.Passback) {
	image, err := i.NewImage()
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		return
	}
	err = ValidateImageParameters(&image)
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		return
	}
	err = StoreImage(image)
	if err == nil {
		id := storage.ImageID{ImageID: image.ImageID}
		pb = model.BuildSuccessPassback(http.StatusOK, id)
	} else {
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
	}
	return
}

// GetImages - returns all images
func GetImages() (pb model.Passback) {
	images := presentation.Images{[]presentation.ImageMarshaled{}}
	imgz, err := GetStoredImages()

	if err == nil {
		for _, i := range imgz {
			images.Images = append(images.Images, presentation.ToImageMarshaled(i))
		}
		pb = model.BuildSuccessPassback(http.StatusOK, images)
	} else {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
	}
	return pb
}

func NumImages() (num int, err error) {
	num = 0
	imgz, err := GetStoredImages()
	if err == nil {
		num = len(imgz)
	}
	return
}

// GetImage - returns an image by imageID
func GetImage(imageID uuid.UUID) (pb model.Passback) {
	image, err := GetStoredImage(imageID)
	if err == nil {
		imageMarshaled := presentation.ToImageMarshaled(image)
		pb = model.BuildSuccessPassback(http.StatusOK, imageMarshaled)
	} else {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
	}
	return pb
}

func GetImageStorage(imageID uuid.UUID) (pb model.Passback) {
	image, err := GetStoredImage(imageID)
	if err == nil {
		//		imageMarshaled := presentation.ToImageMarshaled(image)
		pb = model.BuildSuccessPassback(http.StatusOK, image)
	} else {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
	}
	return pb
}

// DeleteImage - deletes an Image
func DeleteImage(imageID uuid.UUID) (pb model.Passback) {
	_, err := GetStoredImage(imageID)
	if err != nil {
		logrus.Error(err)
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
		return pb
	}

	err = DeleteStoredImage(imageID)
	if err == nil {
		pb = model.BuildSuccessPassback(http.StatusNoContent, nil)
		return pb
	}

	pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
	return pb
}

// UpdateImage - create or update an image
func UpdateImage(image presentation.RawImage, imageid uuid.UUID) (pb model.Passback) {
	i, err := image.NewImage()
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		return
	}
	i.ImageID = imageid

	err = ValidateImageParameters(&i)
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		return
	}
	_, err = GetStoredImage(i.ImageID)
	if err == nil {
		//does exist
		err := StoreImage(i)
		if err == nil {
			pb = model.BuildSuccessPassback(http.StatusOK, nil)
			return pb
		}

	} else {
		//No exist
		err := StoreImage(i)
		if err == nil {
			pb = model.BuildSuccessPassback(http.StatusCreated, nil)
			return pb
		}
	}
	pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
	return pb
}
