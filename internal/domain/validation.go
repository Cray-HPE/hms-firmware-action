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
	"errors"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ValidateActionParameters -> will VALIDATE the parameters, and set a few defaults along the way.
// It will check that an xname is a valid form, but not that it exists.  When we generate the operations we will only generate ones we can
// we will flush out the list and return it to users...
//if they ask for a 'valid' xname that doesnt exist, is that an error? at this point I say NO; dont prevent everthing else
//so the question then becomes, how does an admin learn that an xname wasnt actually used?
func ValidateActionParameters(l *storage.ActionParameters) (err error) {
	if err = ValidateStateComponentFilter(&l.StateComponentFilter); err != nil {
		return err
	}

	if err = ValidateImageFilter(&l.ImageFilter); err != nil {
		return err
	}

	//Tactical decision, not sure if we should do this herem but we are:  basically if they have set an imageID filter;
	//  then override whatever is in version to be explicit; that way we set the ToImage correctly
	if l.ImageFilter.ImageID != uuid.Nil {
		l.Command.Version = "explicit"
	}

	if err = ValidateCommandParameter(&l.Command); err != nil {
		return err
	}

	return nil
}

func ValidateSnapshotParameters(l *storage.SnapshotParameters) (err error) {
	if err = ValidateStateComponentFilter(&l.StateComponentFilter); err != nil {
		return err
	}

	if len(l.Name) == 0 {
		err = errors.New("Name cannot be empty")
		return err
	}

	return nil
}

func ValidateXnames(c *storage.StateComponentFilter) (err error) {
	if len(c.Xnames) > 0 {
		_, badXnames := base.ValidateCompIDs(c.Xnames, false)
		if len(badXnames) > 0 {
			err = model.NewInvalidInputError("invalid/duplicate xnames", badXnames)
			logrus.Error(err)
			return err
		}
	}
	return nil
}

func ValidateStateComponentFilter(c *storage.StateComponentFilter) (err error) {

	if err = ValidateXnames(c); err != nil {
		logrus.Error(err)
		return err
	}

	if len(c.Groups) > 1 && len(c.Partitions) > 1 {
		err = errors.New("illegal, may not have both paritions and groups count > 1")
		logrus.Error(err)
		return err
	}
	return nil
}

func ValidateCommandParameter(c *storage.Command) (err error) {
	if c.Version == "" { //set it to latest if not set.
		c.Version = "latest"
	}

	if c.Tag == "" {
		c.Tag = "default"
	}

	if c.Version != "earliest" && c.Version != "latest" && c.Version != "explicit" {
		err = errors.New("version must be 'earliest' or 'latest'; or you must supply an ImageID")
		logrus.Error(err)
	}
	// at this point there is nothing else to really validate... strings are ""; ints a 0; and bools are (false?)
	return err
}

// ValidateImageFilter - makes sure that if set 1 and ONLY 1 image corresponds to the
func ValidateImageFilter(i *storage.ImageFilter) (err error) {
	if i.ImageID == uuid.Nil {
		return nil
	} else {
		_, err := GetStoredImage(i.ImageID)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}

// Require:
//	ImageID - Non-nil
//	CreateTime -
//	DeviceType - Required
//	Manufacturer - Required
//	Models - Required
//	Target - Required
//	Tags -
//	FirmwareVersion - Required
//	SemanticFirmwareVersion - Required
//	UpdateURI -
//	VersionURI -
//	NeedManualReboot -
//	S3URL - Required
//	AllowableDeviceStates -
//	DependsOn -
func ValidateImageParameters(i *storage.Image) (err error) {
	err = nil
	if i.ImageID == uuid.Nil {
		return errors.New("imageID cannot be Nil")
	}
	if len(i.SoftwareIds) == 0 {
		if len(i.DeviceType) == 0 {
			return errors.New("deviceType is required")
		}
		if len(i.Manufacturer) == 0 {
			return errors.New("manufacturer is required")
		}
		if len(i.Models) == 0 {
			return errors.New("models is required")
		}
	}
	if len(i.Target) == 0 {
		return errors.New("target is required")
	}
	if len(i.FirmwareVersion) == 0 {
		return errors.New("firmwareVersion is required")
	}
	if len(i.S3URL) == 0 {
		return errors.New("S3URL is required")
	}
	if len(i.SemanticFirmwareVersion.String()) == 0 {
		return errors.New("semanticFirmwareVersion is required")
	}
	if len(i.Tags) == 0 {
		return errors.New("tags cannot be empty")
	}

	// TODO: Do we need to check for polling speed?

	return
}
