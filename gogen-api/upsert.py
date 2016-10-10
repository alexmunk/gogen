from __future__ import print_function

import boto3
import json
import httplib
from boto3.dynamodb.conditions import Key, Attr

print('Loading function')


def respond(err, res=None):
    return {
        'statusCode': '400' if err else '200',
        'body': err.message if err else json.dumps(res),
        'headers': {
            'Content-Type': 'application/json',
        },
    }


def lambda_handler(event, context):
    print("Received event: " + json.dumps(event, indent=2))
    body = json.loads(event['body'])
    headers = { }
    if 'Authorization' not in event['headers']:
        return respond(Exception("Authorization header not present"))
    headers['Authorization'] = event['headers']['Authorization']
    headers['User-Agent'] = 'gogen lambda'
    headers['Content-Length'] = 0
    conn = httplib.HTTPSConnection('api.github.com')
    conn.request("GET", "/user", None, headers)
    response = conn.getresponse()
    if response.status != 200:
        return respond(Exception("Unable to authenticate user to GitHub, status: %d, msg: %s, data: %s" % (response.status, response.reason, response.read())))
    validatedbody = { }
    for k, v in body.iteritems():
        if v != "":
            validatedbody[k] = v
    print("Item: ",validatedbody)
    table = boto3.resource('dynamodb').Table('gogen')
    response = table.put_item(
        Item=validatedbody
    )
    return respond(None, response) 