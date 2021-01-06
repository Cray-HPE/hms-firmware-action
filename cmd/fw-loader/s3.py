#  MIT License
#
#  (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
#
#  Permission is hereby granted, free of charge, to any person obtaining a
#  copy of this software and associated documentation files (the "Software"),
#  to deal in the Software without restriction, including without limitation
#  the rights to use, copy, modify, merge, publish, distribute, sublicense,
#  and/or sell copies of the Software, and to permit persons to whom the
#  Software is furnished to do so, subject to the following conditions:
#
#  The above copyright notice and this permission notice shall be included
#  in all copies or substantial portions of the Software.
#
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
#  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
#  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
#  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
#  OTHER DEALINGS IN THE SOFTWARE.

import logging
import boto3
import sys

access_key = ""
secret_key = ""
endpoint = ""
bucket = ""

class client(object):
    def __init__(self, endpt, akey, skey, bckt):
        self.endpoint = endpt
        self.access_key = akey
        self.secret_key = skey
        self.bucket = bckt
        self.resource = boto3.resource(
            's3',
            aws_access_key_id=akey,
            aws_secret_access_key=skey,
            endpoint_url=endpt,
            region_name='',
            verify=False,
        )
        self.cl = boto3.client(
            's3',
            aws_access_key_id=akey,
            aws_secret_access_key=skey,
            endpoint_url=endpt,
            region_name='',
            verify=False,
        )
        try:
            self.cl.head_bucket(Bucket=self.bucket)
        except botocore.exceptions.ClientError:
            self.cl.create_bucket(Bucket=self.bucket)
        # If connecting fails, we throw an exception.
    def test_connection(self):
        try:
            self.cl.head_bucket(Bucket=self.bucket)
        except Exception as err:
            return "S3 Bucket %s does not exist. Error: %s" % (self.bucket, err)
        return None
    def upload_image(self, fp, key, image_data):
        ret = "s3:/%s/%s" % (self.bucket, key)
        try:
            self.cl.head_object(Bucket=self.bucket, Key=key)
            # Object already exists.  Return "URL".  Note that we could at this
            # point verify the new image and the existing one matches, or
            # something like that.  Right now, we are assuming the same key
            # refers to the same file.
            return ret
        except:
            pass
        try:
            metadata = {"imageData":str(image_data)}
            logging.info("Uploading %s", key)
            logging.info("Metadata %s", metadata)
            self.cl.upload_fileobj(fp, self.bucket, key,
                        { "Metadata": metadata , "ACL":"public-read"})
            #https://boto3.amazonaws.com/v1/documentation/api/latest/guide/s3-uploading-files.html#the-extraargs-parameter
        except Exception as e:
            # Failed to upload file to S3
            logging.error("Failed to upload file to S3: %s reason %s", key, e)
            ret = None
        return ret
    def update_image_acl(self, key):
        try:
            logging.info("update ACL to public-read for %s", key)
            #self.cl.put_object_acl({"ACL":"public-read","Bucket":self.bucket, "Key":key})
            #self.cl.put_object_acl(ACL='public-read',Bucket=self.bucket, Key=key)
            object_acl = self.resource.ObjectAcl(self.bucket, key)
            response = object_acl.put(ACL='public-read')
            return True
            #https://boto3.amazonaws.com/v1/documentation/api/latest/guide/s3-uploading-files.html#the-extraargs-parameter
        except Exception as e:
            # Failed to upload file to S3
            logging.error("Failed to upload file to S3: %s reason %s", key, e)
            return False