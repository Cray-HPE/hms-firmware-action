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

package hsm

import (
	"errors"
	"fmt"
	"stash.us.cray.com/HMS/hms-compcredentials"
	"stash.us.cray.com/HMS/hms-securestorage"
	"time"
)

func setupVault(b *HSMv0) (credentials *compcredentials.CompCredStore, err error) {
	for *b.HSMGlobals.Running {
		// StartTime a connection to Vault
		if secureStorage, err := securestorage.NewVaultAdapter(""); err != nil {
			b.HSMGlobals.Logger.Errorf("Unable to connect to Vault, err: %s! Trying again in 1 second...\n", err)
			time.Sleep(1 * time.Second)
		} else {
			b.HSMGlobals.Logger.Infof("Connected to Vault.\n")

			credentials = compcredentials.NewCompCredStore(b.HSMGlobals.VaultKeypath, secureStorage)
			return credentials, nil
		}
	}
	return nil, errors.New("Not running, couldnt connect to vault")
}

func updateHsmDataWithCredentials(b *HSMv0, data *HsmData) (err error) {
	credentials, credErr := b.HSMGlobals.Credentials.GetCompCred(data.ID)
	if credErr != nil {
		err = fmt.Errorf("unable to get credentials for HsmData %s: %s", data.ID, credErr)
		b.HSMGlobals.Logger.Error(err)
		return
	}

	data.User = credentials.Username
	data.Password = credentials.Password

	return
}
