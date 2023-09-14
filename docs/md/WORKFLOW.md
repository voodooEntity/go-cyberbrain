## Workflow

### Internals
go-cyberbrain provides a data driven approach on structuring applications. Rather than defining your application by building a lot of hard chained functions, go-cyberbrain encourages you to split your code into actions representing a specific task.

Each task also defines its requirements to be run. Same as on a function call these requirements are the parameters (data) that is necessary to run. Those actions are developed in the form of plugins.

One of those actions will include the code and also the requirement information. Your final application will be a set of plugins which combined and run with cyberbrain will assemble your application. 

go-cyberbrain by itself will start a set of worker threads called runners. Each of those runners works as executor and supervisor at the same time. This way we avoid the bottleneck of a centralized supervisor. 

As data storage go-cyberbrain uses [gits] which is a thread safe in-memory graph storage i developed. While being very fast, the graph storage allows us to create on dynamic interlinked hive of data. 

When starting go-cyberbrain the application will register the available plugins and start the previously mentioned runners. When registering the plugins cyberbrain will also recognize the execution requirements for each plugin.

### Execution flow
Since go-cyberbrain is a purely data-driven, to start processing data needs to be provided to the running application. This can be done using the [REST api]. 

When new information is passed to the application, it will be checked if a similar information(structure) already exists in the hive. Depending on where from new information is mapped, different strategies may apply. After checking, the application will map those a newly defined information into the graph storage. 

After mapping this information go-cyberbrain will check if the newly mapped data will satisfy the input requirements of any of the registered plugins. This will be done be using the newly mapped data, but can also take in account previously existing data which has been mapped to the newly learned.

Every possible action execution found this way will be mapped into the graph storage as a new job. This way new learned information will always directly be mapped to new jobs and there is no issue with multiple runners dispatching the same job based on identical payloads.

A runner which has no active task is running in a loop and checking if there is any job to apply. As soon a runner finds an unassigned job, it will immediately try to assign it, and if successfully execute it.

On execution the runner will load the necessary action plugin and provide it with the defined input payload. The plugin now can execute its logic without having to care about the application workflow.

Once a plugin execution finishes, a plugin can either just end without returning new data, or the plugin can take the as input provided data and return an enriched version of it. 

If an enriched version of the data is returned, go-cyberbrain will analyze the new data the same way as previously described, map new data into the hive and check & create new jobs based on the newly learned data.

This way your application will always parallelize your workload to the maximum without having to care about issues as race conditions, concurrency and the other fun stuff.

### Example

These tasks 

The possibilities of what you can build with go-cyberbrain are just limited by your imagination, tho for our example on explaining the workflow of go-cyberbrain we are going build a simple webcrawler.

A webcrawler by its nature is a data driven application.