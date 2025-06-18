## preview
![image](https://github.com/user-attachments/assets/3a5064bf-d6f2-429d-b1bb-85c9e271332c)


docker run --name yoru-postgres \
  -e POSTGRES_USER=ryu \
  -e POSTGRES_PASSWORD=pass \
  -e POSTGRES_DB=yoru_pastebin \
  -p 5432:5432 \
  -d postgres:16-alpine



## Contributing

Contributions are welcome! If you have suggestions for improvements, bug fixes, or new features, please feel free to contribute.

## License

This project is licensed under the [**GNU General Public License v3.0 (GPLv3)**](https://www.gnu.org/licenses/gpl-3.0.html).
