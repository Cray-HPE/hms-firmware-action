/*
 * MIT License
 *
 * (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
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
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
)

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
	err = (*GLOB.DSP).StoreImage(image)
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
	imgz, err := (*GLOB.DSP).GetImages()

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
	imgz, err := (*GLOB.DSP).GetImages()
	if err == nil {
		num = len(imgz)
	}
	return
}

// GetImage - returns an image by imageID
func GetImage(imageID uuid.UUID) (pb model.Passback) {
	image, err := (*GLOB.DSP).GetImage(imageID)
	if err == nil {
		imageMarshaled := presentation.ToImageMarshaled(image)
		pb = model.BuildSuccessPassback(http.StatusOK, imageMarshaled)
	} else {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
	}
	return pb
}

func GetImageStorage(imageID uuid.UUID) (pb model.Passback) {
	image, err := (*GLOB.DSP).GetImage(imageID)
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
	_, err := (*GLOB.DSP).GetImage(imageID)
	if err != nil {
		logrus.Error(err)
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
		return pb
	}

	err = (*GLOB.DSP).DeleteImage(imageID)
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
	_, err = (*GLOB.DSP).GetImage(i.ImageID)
	if err == nil {
		//does exist
		err := (*GLOB.DSP).StoreImage(i)
		if err == nil {
			pb = model.BuildSuccessPassback(http.StatusOK, nil)
			return pb
		}

	} else {
		//No exist
		err := (*GLOB.DSP).StoreImage(i)
		if err == nil {
			pb = model.BuildSuccessPassback(http.StatusCreated, nil)
			return pb
		}
	}
	pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
	return pb
}
