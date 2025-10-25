from nacl import public
from nacl.public import PrivateKey, PublicKey, Box
from nacl.signing import SigningKey, VerifyKey
from nacl.encoding import Base64Encoder
from nacl.secret import SecretBox

from argon2.low_level import hash_secret_raw, Type

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

URL = 'https://skam.su:10000'
URL2='wss://skam.su:10000'
sesson = PromptSession()
KDF_TIME = 2
KDF_MEMORY_KB = 64 * 1024
KDF_LEN = 32
KDF_PARALLELISM = 2
SALT_LEN = 16

def derive_key(password: str, salt: bytes):
    pw = password.encode('utf-8')
    key = hash_secret_raw(
        secret=pw,
        salt=salt,
        time_cost=KDF_TIME,
        memory_cost=KDF_MEMORY_KB,
        parallelism=KDF_PARALLELISM,
        hash_len=KDF_LEN,
        type=Type.ID)
    return key

def encrypt_key(priv_pytes: bytes,sign_bytes: bytes, password: str):
    salt = os.urandom(SALT_LEN)
    key = derive_key(password, salt)
    box = SecretBox(key)
    ch_priv = box.encrypt(priv_pytes)
    ch_sign = box.encrypt(sign_bytes)
    payload = {
        "salt": base64.b64encode(salt).decode("ascii"),
        "private_key": base64.b64encode(ch_priv).decode("ascii"),
        "signing_key": base64.b64encode(ch_sign).decode("ascii")}
    return json.dumps(payload)

def decrypt_key(payload_json: str, password: str):
    obj = json.loads(payload_json)
    salt = base64.b64decode(obj['salt'])
    ch_priv = base64.b64decode(obj['private_key'])
    ch_sign = base64.b64decode(obj['signing_key'])
    key = derive_key(password, salt)
    box = SecretBox(key)
    priv_bytes = box.decrypt(ch_priv)
    sign_bytes = box.decrypt(ch_sign)
    return priv_bytes, sign_bytes

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
    print('Сервер отвалился ненадого. Как все заработает, чат восстановится')
    
def on_close(ws, close_status_code, close_msg):
    print('###closed###')
    input('Нажмите Enter чтобы закрыть чат')
    clear_con()
    
def on_open(ws):
    clear_con()
    global name, target_id, priv_key, public_target
    resp = requests.post(f'{URL}/setactive', json={'token':token,
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
        resp = requests.post(f'{URL}/register', json={'name':name, 
                                                                         'public_key':base64.b64encode(pub_key.encode()).decode(),
                                                                         'verify_key':base64.b64encode(verif_key.encode()).decode()})
        if resp.json().get('status') == 'ok':
            token = resp.json().get('token')
            user_id = resp.json().get('id')
            paswd = input('Введите пароль для шифрования ключей. Его нельзя будет восстановить\n>')
            payload = encrypt_key(priv_key.encode(), sign_key.encode(), paswd)
            with open('session.key','w') as f:
                f.write(payload)
            return token, user_id, priv_key, pub_key, sign_key, verif_key, name
        else:
            print('Ошибка при регистрации')
            return register()
    if os.path.exists('session.key') and os.path.getsize != 0:
        while True:
            try: 
                paswd = input('Введите пароль\n>')
                with open('session.key','r') as f:
                    payload = f.read()
                priv_bytes, sign_bytes = decrypt_key(payload, paswd)
                priv_key = PrivateKey(priv_bytes)
                pub_key = priv_key.public_key
                sign_key = SigningKey(sign_bytes)
                verif_key = sign_key.verify_key
                break
            except:
                print('Неверный пароль или повреждённый файл c ключами')
        
        resp = requests.post(f'{URL}/auth-request', json={"public_key": base64.b64encode(pub_key.encode()).decode()})
        if resp.json().get('status') == 'ok':
            seed = resp.json().get('seed')
            signed_seed = sign_key.sign(seed.encode())
            signature = signed_seed.signature
            signed_message = signed_seed.message
            
            resp = requests.post(f'{URL}/auth-verify', json={"signed_message":base64.b64encode(signed_message).decode(),
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
            import_key()
            clear_con()
            return auth()
        
def remove_friend(token: str, tid: int):
    resp = requests.post(f'{URL}/removefriend', json={'token': token, 'target_id': tid})
    if resp.json().get('status') == 'ok':
        print('Друг больше не друг')
        input('Нажмите Enter для выхода')
        return 0
    else:
        print('Что-то пошло не так')
        input('Нажмите Enter для выхода')
        return 0

def remove_chat(token: str, tid: int):
    choice = int(input('''Уверенны, что хотите удалить чат?
[1] Не удалять
[14] Удалить (Навсегда)
> '''))
    if choice == 1:
        return 0
    elif choice == 14:
        
        resp = requests.post(f'{URL}/removechat', json={'token':token, 'target_id': tid})
        
def remove_profile(token: str):
    print('Вы уверенны? Это навсегда')
    choice = int(input('''[1] Нет
[32] Да, удали всё
> '''))
    if choice == 32:
        resp = requests.post(f'{URL}/remove', json={'token':token})
        if resp.json().get('status') == 'ok':
            os.remove('session.key')
            
            input('Всё удалено. Enter для выхода')
            os._exit(0)
        else:
            print('Что-то пошло не так. Надеюсь, все будет работать, хы')

def get_friends(token: str):
    clear_con()
    resp = requests.post(f'{URL}/friends', json={'token':token})
    if resp.json().get('status') == 'ok':
        friends = resp.json().get('friends')
        print('Ваши друзья')
        for i in range(len(friends)):
            print(f'[{i+1}] {friends[i].get('nickname')}')
        print('\n[0] Dыход')
        
        while True:
            choice = int(input('> '))
            if choice<=len(friends) and choice>0:
                return int(friends[choice-1].get('friend_id'))
            elif choice == 0 or choice == 00:
                return 0
            else:
                print('Таких нету')
    elif resp.json().get('status') == 'lonely':
        print('У тебя нет друзей :(')
        return int(input('Можешь ввести id вручную (00 для выхода) > '))
    else:
        print('Что-то пошло не так')
        return int(input('Можешь попробовать ввести id вручную (00 для выхода) > '))

def search_friend():
    clear_con()
    name = input('Введите имя для поиска\n> ')
    resp = requests.post(f'{URL}/search', json={'name':name})
    if resp.json().get('status') == 'ok':
        users = resp.json().get('users')
        print('Найденные пользователи')
        for i in range(len(users)):
            print(f'[{i+1}] {users[i].get('nickname')}')
        print('\n[0] Выход')
        while True:
            choice = int(input('> '))
            if choice<=len(users) and choice>0:
                return int(users[choice-1].get('id'))
            elif choice == 0 or choice == 00:
                return 0
            else:
                print('Нормально напиши')
    else:
        print('Ничего не найдено')
        input('Введите id вручную > ')

def add_friend(token: str):
    global friend_id
    clear_con()
    choice = int(input('[1] Поиск по имени\n[2] Ввести id вручную\n[0] Выйти\n> '))
    if choice == 1:
        friend_id = search_friend()
    elif choice == 0:
        return 0
    elif choice == 2:
        friend_id = int(input('Введи id друга > '))
    if friend_id == 0:
        return 0
    resp = requests.post(f'{URL}/addfriend', json={'token':token,'friend_id': friend_id})
    print('как я тут оказался c',friend_id)
    if resp.json().get('status') == 'ok':
        print('Друг успешно добавлен')
        input('Нажми Enter для выхода')
    elif resp.json().get('status') == 'error':
        if resp.json().get('details') == '58':
            print('Пользователь уже у вас в друзьях')
            input('Нажми Enter для выхода')
        elif resp.json().get('details') == '404':
            print('Пользователь с таким id не найден')
            input('Нажми Enter для выхода')
    else:
        print('Что-то пошло не так')
        input('Нажми Enter для выхода')
        
def get_public_key(target_id: int):
    resp = requests.post(f'{URL}/getpublic', json={'target_id': target_id})
    if resp.json().get('status') == 'ok':
        return resp.json().get('public_key')
    else:
        return 0

def load_msgs(token: str):
    global target_id, public_target, priv_key
    public_bytes = PublicKey(base64.b64decode(public_target))
    box = Box(priv_key, public_bytes)
    resp = requests.post(f'{URL}/messages', json={'token':token,
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

def change_password():
    paswd = input('Введите новый пароль\n>')
    priv_bytes = priv_key.encode()
    sign_bytes = sign_key.encode()
    payload = encrypt_key(priv_bytes, sign_bytes, paswd)
    with open('session.key','w') as f:
        f.write(payload)
    print(f'Пароль успешно сменён на {paswd}')
    input('Нажмите Enter для выхода')
def info(token: str):
    global name
    while True:
        clear_con()
        print(f'''Ник: {name}
Id: {user_id}''')
        choice = int(input('''[1] Изменить ник
[2] Сменить пароль                   
[3] Экспортировать ключи в файл
[4] Удалить профиль
[0] Выход в меню                  
>'''))
        if choice == 1:
            clear_con()
            new_name = input('Введите новый ник > ')
            resp = requests.post(f'{URL}/changename', json={'token':token, 'new_name':new_name})
            if resp.json().get('status') == 'ok':
                name = new_name
                print(f'Ваш новый ник : {name}')
                input('Нажмите Enter')
            else:
                print('Что-то пошло не так...')
                input('Нажмите Enter для выхода')
        elif choice == 2:
            change_password()
        elif choice == 3:
            export_key()
        elif choice == 4:
            remove_profile(token)
        elif choice == 0:
            break

def actions(target_id: int):
    while True:
        clear_con()
        choice = int(input('''Действия
[1] Открыть чат
[2] Удалить друга
[3] Удалить чат
[0] Назад                           
> '''))
        if choice == 1:
            start_chat(token)
        elif choice == 2:
            remove_friend(token, target_id)
            break
        elif choice == 3:
            remove_chat(token, target_id)
        elif choice == 0:
            break

def import_key():
    def skuka(path: str):
        with open(path,'r') as f:
            lines = f.readlines()
        priv_bytes = base64.b64decode(lines[0])
        sign_bytes = base64.b64decode(lines[1])
        paswd = input('Введите пароль для шифрования ключей. Восстановить его не получится\n>')
        payload = encrypt_key(priv_bytes, sign_bytes, paswd)
        with open('session.key','w') as f:
            f.write(payload)
        priv_key = PrivateKey(priv_bytes)
        pub_key = priv_key.public_key
        sign_key = SigningKey(sign_bytes)
        verif_key = sign_key.verify_key
        return 0
    clear_con()
    if os.path.exists('keys.txt') and os.path.getsize('keys.txt') != 0:
        print('Обнаружен файл с ключами. Использовать его?')
        choice = int(input('[1] Да\n[2] Нет\n> '))
        if choice == 1:
            skuka('keys.txt')
            return 0
        
    while True:
        print('Введите путь до файла с ключами')
        path = input('> ')
        try:
            skuka(path)
            break
        except:
            print('Неверный путь')
    return 0

def export_key():
    priv = base64.b64encode(priv_key.encode()).decode()+'\n'
    sign = base64.b64encode(sign_key.encode()).decode()
    with open('keys.txt','w') as f:
        f.write(priv)
        f.write(sign)
    print('Ключи успешно записаны в файл keys.txt. Файл находится в корне программы.')
    input('Нажминет Enter для выхода')

def start_chat(token:str):
    
    websocket.enableTrace(False)
    ws = websocket.WebSocketApp(f'{URL2}/ws?token={token}',
                                    on_open=on_open,
                                    on_message=on_message,
                                    on_error=on_error,
                                    on_close=on_close)
    ws.run_forever(dispatcher=None, reconnect=5)


if __name__ == '__main__':
    print('Ждём пока сервер проснётся. Может занять несколько минут')

    token, user_id, priv_key, pub_key, sign_key, verif_key, name = auth()
    while True:
        clear_con()
        choice = int(input('''[1] Список друзей
[2] Добавить друга
[3] Информация                           
[0] Выйти\n> '''))
        if choice == 1:
            target_id = get_friends(token)
            if target_id == 0:
                pass
            else:
                public_target = get_public_key(target_id)
                if public_target == 0:
                    print('Такого id нет в базе')
                    input('Нажмите Enter для выхода')
                else:
                    actions(target_id)
        elif choice == 2:
            add_friend(token)
        elif choice == 3:
            info(token)
            
        elif choice == 0:
            break
        else:
            print('Неверный ввод')
        clear_con()
        
