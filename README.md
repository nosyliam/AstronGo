AstronGo
--------------
_A Server Technology for Realtime Object Networking in GoLang_

AstronGo is a rewrite of Astron, an open-source, distributed server suite particularly well suited
for powering MMO games written in C++. Inspired by major issues with Astron, especially stability and diagnostics, it has been completely 
rewritten in GoLang which provides built-in support for networking and concurrency as a language.

For more information about Astron, please visit their page [here](https://github.com/Astron/Astron).

## Roadmap

As of January 2020, the foundation of AstronGo is complete. This includes core components such as
the DClass system, the message director, and their associated unit testing. There are no exact deadlines
for development milestones, however, this is a basic in-order list of components to be written.
1) Eventlogger
2) ClientAgent
    - Configuration Implementation and Testing
    - Haproxy Implementation
3) Client
    - Research and Development
    - Architecture Redesign and Implementation
4) Unit Testing (by June 2020)
5) StateServer
6) Database


## Contributing

Currently, AstronGo is a passion project by Liam (me). If you have any questions about the project
or would like to contribute to it's development please contact me on Discord at **liam#0002**. Thank you :)
