import requests
import json
import time

import core.config as core_config
import core.logger as core_logger

config = core_config.config
logger = core_logger.logger

app_access_token = ''
expire = 0


def send_post_message(title, text, buf=None):
    url = 'https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=chat_id'
    content = [[
        {
            'tag': 'text',
            'text': text
        }
    ]]
    if buf is not None:
        content.append([
            {
                'tag': 'img',
                'image_key': upload_image(buf)
            }
        ])
    content.append([
        {
            'tag': 'at',
            'user_id': 'all'
        }
    ])
    access_token = get_access_token()
    payload = json.dumps({
        'content': json.dumps({
            'zh_cn': {
                'title': title,
                'content': content,
            }
        }),
        'msg_type': 'post',
        'receive_id': config['feishu']['receive_id'],
    })
    headers = {
        'Content-Type': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    res = requests.request('POST', url, headers=headers, data=payload).json()
    if res['code'] != 0:
        logger.error(f'飞书消息发送失败：{res}')


def upload_image(image):
    url = 'https://open.feishu.cn/open-apis/im/v1/images'
    access_token = get_access_token()
    payload = {'image_type': 'message'}
    files = [
        ('image', ('klines.png', image, 'image/png'))
    ]
    headers = {
        'Authorization': f'Bearer {access_token}'
    }

    res = requests.request('POST', url, headers=headers, data=payload, files=files).json()
    if res['code'] != 0:
        logger.error(f'飞书图片上传失败：{res}')

    return res['data']['image_key']


def get_access_token():
    global app_access_token, expire
    # 如果没有过期，直接返回
    if expire > int(time.time()):
        return app_access_token

    url = 'https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal'
    headers = {
        'Content-Type': 'application/json'
    }
    payload = json.dumps({
        'app_id': config['feishu']['app_id'],
        'app_secret': config['feishu']['app_secret']
    })
    res = requests.request('POST', url, headers=headers, data=payload).json()
    if res['code'] != 0:
        logger.error(f'飞书access_token获取失败：{res}')
    else:
        app_access_token = res['app_access_token']
        expire = int(time.time()) + res['expire'] - 60

    return app_access_token
