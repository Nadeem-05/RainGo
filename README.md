# RainGo

RainGo is an open-source password recovery and analysis tool written in Go, leveraging **rainbow tables** for efficient hash cracking. It is packaged as a user-friendly desktop application using the Wails framework.

---
<img src="https://i.ibb.co/nM3gNSw/Screenshot-From-2024-12-12-21-04-02.png" alt="Raingo Desktop" border="0"></img>
## Features

- **Rainbow Table Management**: Generate, store, and use precomputed rainbow tables for various hash algorithms.
- **Cross-Platform**: Pre-built binaries available for Windows, macOS, and Linux.
- **Intuitive Interface**: Simplified interaction through a desktop application.

---

## Usage

Simply install the pre-built binary file available in the releases section for your platform or build it yourself using the instructions below. Run/Install the application as a desktop app to start using RainGo.

---

## Getting Started

### Prerequisites

Ensure you have the following installed on your system:

- [Go](https://go.dev/dl/) (1.20 or later)

### Building

1. Clone the repository:
   ```bash
   git clone https://github.com/Nadeem-05/RainGo.git
   cd RainGo
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Building the application:
   - RainGo uses the Wails framework to package the application. Follow these steps to build for your platform:
     - **Windows**:
       ```bash
       wails build -platform windows/amd64
       ```
     - **macOS**:
       ```bash
       wails build -platform darwin/universal
       ```
     - **Linux**:
       ```bash
       wails build -platform linux/amd64
       ```
     - The pre-built binary files (.exe, .app) are available for installation in the `build/bin` directory.

4. Run the application (Mac/Linux):
   ```bash
   ./raingo
   ```

---

## Contributing

Contributions are welcome! Follow these steps to contribute:

1. Fork the repository.
2. Create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature-name
   ```
3. Commit your changes and push to your fork:
   ```bash
   git commit -m "Add a meaningful commit message"
   git push origin feature-name
   ```
4. Submit a pull request with a detailed description.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Author

- **Nadeem**  
  GitHub: [Nadeem-05](https://github.com/Nadeem-05)  
  Email: nadeem@example.com (replace with actual contact details if applicable)

---

## Acknowledgments

- **RockYou**: For providing the password list.
- **Go Tour**: For the learning experience
- The open-source community for inspiring the development of RainGo.

