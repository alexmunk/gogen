from __future__ import print_function

import boto3
import json
import decimal

from boto3.dynamodb.conditions import Key, Attr

print('Loading function')


def decimal_default(obj):
    if isinstance(obj, decimal.Decimal):
        return float(obj)
    raise TypeError


def respond(err, res=None):
    return {
        'statusCode': '400' if err else '200',
        'body': err.message if err else json.dumps(res, default=decimal_default),
        'headers': {
            'Content-Type': 'application/json',
        },
    }


def lambda_handler(event, context):
    print("Received event: " + json.dumps(event, indent=2))
    q = event['pathParameters']['proxy']
    print("Query: ", q)

    table = boto3.resource('dynamodb').Table('gogen')
    response = table.get_item(Key={"gogen": q})

    # print("Response: " + json.dumps(response, indent=2))
    if 'Item' not in response:
        return {
            'statusCode': '404',
            'body': 'Could not find Gogen: %s' % q,
        }
    if 'gogen' not in response["Item"]:
        return {
            'statusCode': '404',
            'body': 'Could not find Gogen: %s' % q,
        }
    return respond(None, response)
