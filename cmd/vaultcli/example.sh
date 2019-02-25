#!/usr/bin/env sh

# Note: the password or hash must be in a string format, or terminal may not read them correct.

# To hash a password.
vaultcli hash 'my_password'

# To validate a hash passsword pair.
vaultcli validate 'my_hash' 'my_hash'