# IMPORTANT NOTE
go-cyberbrain at this moment is in a pre-alpha experimental state. Im right now working on the docs because without them its quite impossible to dive into it. I will add more documentations in the near future, but due to limited time i can spend on this project i can't provide a final date. 

Im happy if you're interested in the topic and may revisit the project later.

voodooEntity aka laughingman

---------------

## go-cyberbrain

go-cyberbrain is an innovative architectural framework for building intelligent applications. It provides a solid foundation that allows developers to effortlessly construct applications by leveraging custom plugins tailored to their specific requirements.

**Note:** go-cyberbrain is currently in the early alpha state and should be considered more experimental than production-ready. While it showcases the idea of the architecture, further development is underway to reach a production-ready state.

### Features

- Lightweight and optimized: go-cyberbrain is built using mostly standard Go libraries, ensuring a streamlined development experience without the need for third-party dependencies other than created the same developers.
- Self-supervising and orchestrating: With go-cyberbrain, the architecture takes charge of managing the workload and scheduling, allowing developers to focus on developing plugins and enhancing the application's functionality.
- Extensibility through plugins: Plugins define the specific requirements and behavior of the application. By developing custom plugins, developers can extend go-cyberbrain's capabilities and tailor the application to their needs.
- Data-driven processing: go-cyberbrain embraces a data-driven approach, seamlessly handling complex data flows. Its in-memory graph backend facilitates the storage, analysis, and enrichment of interconnected information.
- Multithreaded by design: go-cyberbrain optimizes job execution by utilizing a multi threaded architecture. It ensures efficient processing, avoids duplication of tasks, and maximizes performance.

For more information on how cyberbrain works check the [workflow] page. It provides a more detailed explanation about the workflow itself and why it's different than current standard architectures. 

### Getting Started

To explore the capabilities of go-cyberbrain and start building intelligent applications, follow these steps:

1. Clone go-cyberbrain to a directory of your choice
2. Starting in the project root directory 
```
cd ./cmd/cyberbrain
go build -o cyberbrain && cp cyberbrain ~/go/bin
```
3. Make sure ~/go/bin is in your PATH. You may choose another directory to store the binary which is in your PATH, in that case modify the path given in Step 2. command to your desired destination.
4. Clone the example project repo [url] into a directory of your choice. This will turn into your project.
5. Use the following command to make sure your cyberbrain installation works fine, startin in the example projects root directory
```
cyberbrain test
```
6. If cyberbrain test finishes without errors your installation seems to be fine. If errors occure please check the [troubbleshooting] section.

You now have a working version of cyberbrain setup for development and usage on your system. Since cyberbrain is an application architecture/framework it needs plugins to work. For information on plugin development check [your first project] or [plugin development]


