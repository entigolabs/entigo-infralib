# entigo-infralib


Usually we release once per day. During the evening the entigo-infralib AWS account is nuked(Nuke action). In the morning latest release is installed and tests executed(Stable action) and after that it is upgraded to "main" branch and tests are executed. If the tests are passed and main is not the same state as last release, then a new release is created.

Once the release is created then we make it public in [entigo-infralib-release](https://github.com/entigolabs/entigo-infralib-release) repository.

To create a new release the "Release" action should be run. It will create a release if main branch differs from last release.


## Folders ##

__modules__ contains opinnionated terraform modules or kubernetes helm charts that we repeatedly use in our projects.
__images__ contains the runtime images for running infrastructure as code.
__profiles__ contains base profiles used by entigo-infralib-agent that we use on many clients. It is a way of combining different module with common inputs.
__providers__ contians provider configurations for terraform modules.
