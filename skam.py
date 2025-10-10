from nacl import public
from nacl.public import PrivateKey, PublicKey, Box
from nacl.signing import SigningKey, VerifyKey
from nacl.encoding import Base64Encoder

from prompt_toolkit import PromptSession
from prompt_toolkit.patch_stdout import patch_stdout
import requests

import base64
import sys
import threading
import time
import os
import json
import websocket
import rel


sesson = PromptSession()



def clear_con():
    os.system('cls' if os.name == 'nt' else 'clear')
    
def on_message(ws, message):
    global target_id, public_target, priv_key
    public_bytes = PublicKey(base64.b64decode(public_target))
    box = Box(priv_key, public_bytes)
    data = json.loads(message)
    name = data.get('name')
    snd = data.get('message')
    message = box.decrypt(base64.b64decode(snd))
    message = message.decode("utf-8")
    with patch_stdout():
        print(f'[{name}]: {message}')

    
def on_error(ws, error):
    print(error)
    
def on_close(ws, close_status_code, close_msg):
    print('###closed###')
    input('Press Enter to close chat')
    clear_con()
    
def on_open(ws):
    clear_con()
    global name, target_id, priv_key, public_target
    resp = requests.post(f'https://skam.onrender.com/setactive', json={'token':token,
                                                                      'target_id':target_id})
    public_bytes = PublicKey(base64.b64decode(public_target))
    box = Box(priv_key, public_bytes)
    print('Connected to server. Type /quit to quit')
    load_msgs(token)
    
    def run():
        while True:
            message = sesson.prompt(f'[{name}]: ')
            if message == '/quit':
                ws.close()
                break
            else:
                chiper = box.encrypt(message.encode("utf-8"))
                snd = base64.b64encode(chiper).decode()
                payload = {'target_id':target_id, 'message':snd, 'name':name}
                ws.send(json.dumps(payload))
    threading.Thread(target=run, daemon=True).start()
    
def auth():
    def register():

        name = input('Введите имя > ')
        priv_key = PrivateKey.generate()
        pub_key = priv_key.public_key
        sign_key = SigningKey.generate()
        verif_key = sign_key.verify_key
        resp = requests.post('https://skam.onrender.com/register', json={'name':name, 
                                                                         'public_key':base64.b64encode(pub_key.encode()).decode(),
                                                                         'verify_key':base64.b64encode(verif_key.encode()).decode()})
        if resp.json().get('status') == 'ok':
            token = resp.json().get('token')
            user_id = resp.json().get('id')
            with open('private.key', 'wb') as f:
                f.write(base64.b64encode(priv_key.encode()))
            with open('signing.key', 'wb') as f:
                f.write(base64.b64encode(sign_key.encode()))
            return token, user_id, priv_key, pub_key, sign_key, verif_key, name
        else:
            print('Ошибка при регистрации')
            return register()
        
    if (os.path.exists('private.key') and 
        os.path.getsize('private.key') != 0 and
        os.path.exists('signing.key') and 
        os.path.getsize('signing.key') != 0):
        
        with open('private.key','rb') as f:
            priv_bytes = base64.b64decode(f.read())
            priv_key = PrivateKey(priv_bytes)
            
            pub_key = priv_key.public_key
              
        with open('signing.key','rb') as f:
            sign_bytes = base64.b64decode(f.read())
            sign_key = SigningKey(sign_bytes)
            
            verif_key = sign_key.verify_key
            
        resp = requests.post('https://skam.onrender.com/auth-request', json={"public_key": base64.b64encode(pub_key.encode()).decode()})
        if resp.json().get('status') == 'ok':
            seed = resp.json().get('seed')
            signed_seed = sign_key.sign(seed.encode())
            signature = signed_seed.signature
            signed_message = signed_seed.message
            
            resp = requests.post('https://skam.onrender.com/auth-verify', json={"signed_message":base64.b64encode(signed_message).decode(),
                                                                                "signed_seed": base64.b64encode(signature).decode(),
                                                                                "public_key": base64.b64encode(pub_key.encode()).decode()})
                
            if resp.json().get('status') == 'ok':
                token = resp.json().get('token')
                user_id = resp.json().get('id')
                name = resp.json().get('name')
                return token, user_id, priv_key, pub_key, sign_key, verif_key, name
            else:
                print('Ошибка на сервере. Повтор через 15 секунд')
                time.sleep(15)
                print(resp.json().get('status'))
                return auth()
        else:
            return register()
                

    else:
        choice = int(input('''Файл авторизации не найден
[1] Зарегистрироваться
[2] Импорт ключей из файла                          
> '''))
        if choice == 1:
            return register()
        if choice == 2:
            print('пока не доделал это')
        

def get_friends(token: str):
    clear_con()
    resp = requests.post(f'https://skam.onrender.com/friends', json={'token':token})
    if resp.json().get('status') == 'ok':
        friends = resp.json().get('friends')
        print('Ваши друзья')
        for i in range(len(friends)):
            print(f'[{i+1}] {friends[i].get('nickname')}')
        
        
        while True:
            choice = int(input('Кому хотите написать > '))
            if choice<=len(friends) and choice>=0:
                return int(friends[choice-1].get('friend_id'))
            else:
                print('Нормально напиши, черт')
    elif resp.json().get('status') == 'lonely':
        print('У тебя нет друзей :(')
        return int(input('Можешь ввести id вручную (00 для выхода) > '))
    else:
        print('Что-то пошло не так')
        return int(input('Можешь попробовать ввести id вручную (00 для выхода) > '))
    
def add_friend(token: str):
    clear_con()
    friend_id = int(input('Введи id друга > '))
    resp = requests.post(f'https://skam.onrender.com/addfriend', json={'token':token,
                                                                       'friend_id': friend_id})
    if resp.json().get('status') == 'ok':
        print('Друг успешно добавлен')
        
def get_public_key(target_id: int):
    resp = requests.post(f'https://skam.onrender.com/getpublic', json={'target_id': target_id})
    if resp.json().get('status') == 'ok':
        return resp.json().get('public_key')
    else:
        return 0

def load_msgs(token: str):
    global target_id, public_target, priv_key
    public_bytes = PublicKey(base64.b64decode(public_target))
    box = Box(priv_key, public_bytes)
    resp = requests.post(f'https://skam.onrender.com/messages', json={'token':token,
                                                                      'target_id':target_id})
    if resp.json().get('status') == 'none':
        print('Чат пуст. Напишите что-нибудь')
    elif resp.json().get('status') == 'ok':
        messages = resp.json().get('messages')
        for i in messages:
            snd = i.get('message')
            message = box.decrypt(base64.b64decode(snd))
            message = message.decode("utf-8")
            print(f'[{i.get('name')}]: {message}')
    else:
        print('Возникла ошибка при загрузке истории сообщений')
        
def start_chat(token:str):
    
    websocket.enableTrace(False)
    ws = websocket.WebSocketApp(f'wss://skam.onrender.com/ws?token={token}',
                                    on_open=on_open,
                                    on_message=on_message,
                                    on_error=on_error,
                                    on_close=on_close)
    ws.run_forever(dispatcher=None, reconnect=5)


if __name__ == '__main__':
    

    token, user_id, priv_key, pub_key, sign_key, verif_key, name = auth()
    while True:
        clear_con()
        choice = int(input('''[1] Открыть чат
[2] Добавить друга
[3] Выйти\n> '''))
        if choice == 1:
            target_id = get_friends(token)
            if target_id == 00:
                pass
            else:
                public_target = get_public_key(target_id)
                if public_target == 0:
                    print('Такого id нет в базе')
                    input('Нажмите Enter для выхода')
                else:
                    start_chat(token)
        elif choice == 2:
            add_friend(token)
        elif choice == 3:
            break
        else:
            print('Неверный ввод')
        clear_con()
        