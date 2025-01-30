# passcualito
*Simple Command-Line Password Manager for Linux*

## Caveats
- Go version **1.23.4**
- This library has been developed on and for Linux following open source philosophy.

## Installation
- Downloading, compiling and installing manually:
```bash
git clone https://github.com/javiorfo/passcualito
cd passcualito
sudo make clean install
```

- From AUR Arch Linux:
```bash
yay -S passcualito
```

## Description
```text
Usage:
  passc [command]

Available Commands:
  add         Add a new entry to the store
  completion  Generate the autocompletion script for the specified shell
  copy        Copy password to clipboard
  edit        Edit the entry.
  export      Export data in a JSON file
  help        Help about any command
  import      Import entries from a JSON file
  list        List all properties of the entry by name
  logout      Logout of the app
  password    Generates a password of the number passed
  remove      Remove the entry
  version     app version

Flags:
  -h, --help   help for passc

Use "passc [command] --help" for more information about a command.
```

## Usage
- By executing any command, if there is no password store created, **passcualito** will ask for a `Master Password` (6 characters at least). 
- Once the master password is created also the password store will be (**$HOME/.passcualito/store.passc**)
- When the user is logged, **passcualito** will keep some kind of session using **/tmp system folder**


<img src="https://github.com/javiorfo/img/blob/master/passcualito/passcualito.gif?raw=true" alt="passcualito"/>

#### Notes
- Command `passc add entry_name` could have optionals flags: 
    - **-p p4$$w0rd_here** (if not enter a password manually a random 20 char password will be generated) 
    - **-i "some extra useful info"** 
- Command `passc password 10` (10 char password) could have optionals flags: 
    - **-c a** (alphabetic password)
    - **-c n** (numeric password)
    - **-c an** (alphanumeric password)
    - **-c anc** (alphanumeric + capitals password)
- Format of **JSON** data:
```json
{
  "name": "name_of_entry",
  "password": "password_value",
  "info": "some extra info (could be empty)"
}
```

---

### Donate
- **Bitcoin** [(QR)](https://raw.githubusercontent.com/javiorfo/img/master/crypto/bitcoin.png)  `1GqdJ63RDPE4eJKujHi166FAyigvHu5R7v`
- [Paypal](https://www.paypal.com/donate/?hosted_button_id=FA7SGLSCT2H8G)
