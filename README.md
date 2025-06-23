# Modular MIDI Control - PC Driver

This repository contains the PC driver software for a modular MIDI input device designed for music production, DJing, and lighting control. The software is currently in development and will serve as the central hub for a flexible and expandable hardware ecosystem.

## The Vision: A Modular MIDI Ecosystem

The ultimate goal is to create a versatile MIDI controller system with a central **Main Module** powered by an ESP32. This core unit will feature a variety of built-in controls like sliders, rotary encoders, and buttons.

The Main Module will connect to your computer via either a wired USB-C connection or wirelessly over a Wi-Fi UDP stream. This allows for seamless integration with your favorite Digital Audio Workstation (DAW), virtual DJ software, lighting control programs, or any other application that accepts MIDI input.

The system is designed to be expandable. The Main Module will be equipped with magnetic, spring-loaded pins on its sides, allowing for the easy attachment of various **expansion modules**. Each of these modules will be managed by its own microcontroller (such as an MSP430 or equivalent), which will have a unique ID. Data from these expansion modules will be intelligently relayed through any intermediary modules to the Main Module, creating a robust and scalable control surface.

## Focus of this Repository

**This repository is currently dedicated to the development of the PC driver software.** The hardware components and designs will be added at a later stage.

The PC software is being developed to be cross-platform and will be available for:

* **Linux** (the primary development environment)
* **Windows**
* **macOS**

This driver will also be a valuable tool for anyone looking to create their own DIY ESP32-based MIDI input device. Comprehensive documentation and tutorials will be provided to guide you through the process.

## Transparency

To be transparent about the development process:

* **Artificial Intelligence (AI)** is used to assist in generating UI styles, small code portions and documentation texts (AND REGEX THIS BITCH). It is also used as a tool to better understand complex error messages.
* The core **code and hardware design** are originally created by me, Jorim.

## How to build: 
*(the project isnt in a shipable state yet. Contact me if you want to help with development (stern@jorim.xyz))*


Instructions on how to install and use the driver software will be available here. The aim is to provide a user-friendly experience for both the modular hardware and for those using their own custom ESP32 MIDI devices.

## Contributing

Feel free to help grow this project. If you would like to get started, message me on my socials in my profiles github readme :) 

---

Developed by **Jorim**.
