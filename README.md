# Modular MIDI Control - PC Driver

This repository contains the PC driver software for a modular MIDI input device designed for music production, DJing, and lighting control. The software is currently in development and will serve as the central hub for a flexible and expandable hardware ecosystem.

## The Vision: A Modular MIDI Ecosystem

The ultimate goal is to create a versatile MIDI controller system with a central **Main Module** powered by an ESP32. This core unit will feature a variety of built-in controls, such as sliders, rotary encoders, and buttons.

The Main Module will connect to a computer via either a wired USB-C connection or a wireless UDP stream, allowing for seamless integration with your favorite Digital Audio Workstation (DAW), virtual DJ software, lighting control programs, or any other application that accepts MIDI input.

The system is designed to be expandable. The Main Module will be equipped with magnetic, spring-loaded pins on its sides, allowing for the easy attachment of various **expansion modules**. Each of these modules will be managed by its own microcontroller (such as an MSP430 or equivalent) with a unique ID. Data from the expansion modules will be intelligently relayed through any intermediary modules to the Main Module, creating a robust and scalable control surface.

## Focus of this Repository

**This repository is dedicated to the development of the PC driver software.** The hardware components and designs will be added at a later stage.

The PC software is being developed to be cross-platform and will be available for:

* **Linux** (the primary development environment)
* **Windows**
* **macOS**

This driver will also be a valuable tool for anyone looking to create their own DIY ESP32-based MIDI input device. Comprehensive documentation and tutorials will be provided to guide you through the process.

## Transparency

To be transparent about the development process:

* **Artificial Intelligence (AI)** is used to assist in generating UI styles, small code portions, and documentation. It is also used as a tool to better understand complex error messages.
* The core **code and hardware design** are originally created by me, Jorim.

## Getting Started

### Install

Coming soon.

### How to Build

To build the project, you need:

* **Go** (Golang) installed on your system.
* The Go modules specified in the **[go.mod](https://github.com/jrmfrbg/modularMidiDriver/blob/0d215bba0c013447b421377cb13d739b448aa301/go.mod)** file.

1.  Clone this repository.
2.  Navigate to the repository folder.
3.  Run the following commands:
    ```bash
    go build -o frontend/frontend frontend/main.go
    go build -o backend/driver/backend backend/driver/main.go
    ```
4.  You have now generated the executables (`frontend/frontend` and `backend/driver/backend`). On Windows, these will be `.exe` files. First, run the backend executable. After that, navigate to the frontend directory and run the frontend executable with the required parameters. To see available parameters, use `./frontend help`.

---

Developed by **[Jorim](https://instagram.com/jrm.frbg)**.
