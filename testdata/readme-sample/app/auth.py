import hashlib


def legacy_token(user):
    return hashlib.md5(user.encode()).hexdigest()
