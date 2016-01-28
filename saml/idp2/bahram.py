import json
import scrypt
import base64

from etcd import Client as ETCDClient
from etcd import EtcdKeyNotFound


class BahramUserManager(object):
    table_name = 'bahram/users/'
    etcd_client = None
    BAHRAM_PASSWORD_SALT = 'xk3HY0Yrg9fUKMdOFqdmiYdLCsCX6WuLYlC/fqC1VKw='

    def __init__(self, etcd_host, etcd_port):
        self.etcd_client = ETCDClient(host=etcd_host, port=etcd_port)

    @staticmethod
    def encrypt(password):
        return scrypt.hash(password=password, salt=base64.decodestring(BahramUserManager.BAHRAM_PASSWORD_SALT),
                           N=16384, r=8, p=1, buflen=32)

    def verify_email_and_password(self, email, password):
        try:
            user_json = self.etcd_client.get(self.table_name + email).value
            loaded_user = json.loads(user_json)
            user_decoded_password = base64.decodestring(str(loaded_user.get("password")))
            hashed_password = self.encrypt(str(password))
            return user_decoded_password == hashed_password
        except EtcdKeyNotFound:
            return False
