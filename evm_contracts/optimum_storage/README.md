```bash
yay -S nvm
echo 'export NVM_DIR="$HOME/.nvm"' >> ~/.zshrc
echo '[ -s "/usr/share/nvm/init-nvm.sh" ] && source /usr/share/nvm/init-nvm.sh' >> ~/.zshrc
source ~/.zshrc

# install & use Node 22 (LTS)
nvm install 22
nvm use 22
```
