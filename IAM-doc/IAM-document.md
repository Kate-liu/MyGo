## IAM系统概述

### 项⽬背景：为什么选择 IAM 系统作为实战项⽬？

在做 Go 项⽬开发时，绕不开的⼀个话题是安全，如何保证 Go 应⽤的安全，是每个开发者都要解决的问题。虽然 Go 应⽤的安全包含很多⽅⾯，但⼤体可分为如下 2 类：

- 服务⾃身的安全：为了保证服务的安全，需要禁⽌⾮法⽤户访问服务。这可以通过服务器层⾯和软件层⾯来解决。服务器层⾯可以通过物理隔离、⽹络隔离、防⽕墙等技术从底层保证服务的安全性，软件层⾯可以通过 HTTPS、⽤户认证等⼿段来加强服务的安全性。
- 服务器层⾯⼀般由运维团队来保障，软件层⾯则需要开发者来保障。
  服务资源的安全：服务内有很多资源，为了避免⾮法访问，开发者要避免 UserA 访问到UserB 的资源，也即需要对资源进⾏授权。通常，可以通过资源授权系统来对资源进⾏授权。

总的来说，为了保障 Go 应⽤的安全，需要对访问进⾏认证，对资源进⾏授权。那么，要如何实现访问认证和资源授权呢？

认证功能不复杂，可以通过 JWT （JSON Web Token）认证来实现。授权功能⽐较复杂，授权功能的复杂性使得它可以囊括很多 Go 开发技能点。因此，将认证和授权的功能实现升级为 IAM 系统，通过讲解它的构建过程，讲清楚 Go 项⽬开发的全部流程。



### IAM 系统是什么？

IAM（Identity and Access Management，身份识别与访问管理）系统是⽤ Go 语⾔编写的⼀个 Web 服务，⽤于给第三⽅⽤户提供访问控制服务。

IAM 系统可以帮⽤户解决的问题是： 在特定的条件下，谁能够 / 不能够对哪些资源做哪些操作（ Who is able to do what on something given some context），也即完成资源授权功能。

> 在提到 IAM 系统或者 IAM 时都是指代 IAM 应⽤。

那么，IAM 系统是如何进⾏资源授权的呢？

下⾯，通过 IAM 系统的资源授权的流程，来看下它是如何⼯作的，整个过程可以分为 4 步。

![image-20211103224203448](IAM-document.assets/image-20211103224203448.png)

1. ⽤户需要提供昵称、密码、邮箱、电话等信息注册并登录到 IAM 系统，这⾥是以⽤户名和密码作为唯⼀的身份标识来访问 IAM 系统，并且完成认证。
2. 因为访问 IAM 的资源授权接⼝是通过密钥（secretID/secretKey）的⽅式进⾏认证的，所以⽤户需要在 IAM 中创建属于⾃⼰的密钥资源。
3. 因为 IAM 通过授权策略完成授权，所以⽤户需要在 IAM 中创建授权策略。
4. 请求 IAM 提供的授权接⼝，IAM 会根据⽤户的请求内容和授权策略来决定⼀个授权请求是否被允许。

可以看到，在上⾯的流程中，IAM 使⽤到了 3 种系统资源：⽤户（User）、密钥（Secret）和策略（Policy），它们映射到程序设计中就是 3 种 RESTful 资源：

- ⽤户（User）：实现对⽤户的增、删、改、查、修改密码、批量修改等操作。
- 密钥（Secret）：实现对密钥的增、删、改、查操作。
- 策略（Policy）：实现对策略的增、删、改、查、批量删除操作。



### IAM 系统的架构⻓啥样？

知道了 IAM 的功能之后，再来详细说说 IAM 系统的架构，架构图如下：

![image-20211103224631181](IAM-document.assets/image-20211103224631181.png)

总的来说，IAM 架构中包括 9 大组件和 3 大数据库。

将这些组件和功能都总结在下面的表格中。这里面，主要记住 5 个核心组件，包括 iam-apiserver、iam-authz-server、iam-pump、marmotedu-sdk-go 和 iamctl 的功能，还有 3 个数据库 Redis、MySQL（MariaDB） 和 MongoDB 的功能。

![image-20211103225006419](IAM-document.assets/image-20211103225006419.png)

> 前 5 个组件是需要实现的核心组件。 
>
> 后 4 个组件是一些旁路组件，不影响项目的使用。 如果感兴趣，可以自行实现。

此外，IAM 系统为存储数据使用到的 3 种数据库的说明如下所示。

![image-20211103225330420](IAM-document.assets/image-20211103225330420.png)



### 通过使用流程理解架构 

只看到这样的系统架构图和核心功能讲解，可能还不清楚整个系统是如何协作，来最终完成资源授权的。所以接下来，通过详细讲解 IAM 系统的使用流程及其实现细节，来进一步加深对 IAM 架构的理解。总的来说，可以通过 4 步去使用 IAM 系统的核心功能。 

#### 第 1 步，创建平台资源

用户通过 iam-webconsole（RESTful API）或 iamctl（sdk marmotedu-sdk-go）客户端请求 iam-apiserver 提供的 RESTful API 接口完成用户、密钥、授权策略的增删改查， iam-apiserver 会将这些资源数据持久化存储在 MySQL 数据库中。而且，为了确保通信安全，客服端访问服务端都是通过 HTTPS 协议来访问的。 

#### 第 2 步，请求 API 完成资源授权

用户可以通过请求 iam-authz-server 提供的 /v1/authz 接口进行资源授权，请求 /v1/authz 接口需要通过密钥认证，认证通过后 /v1/authz 接口会查询授权策略，从而决定资源请求是否被允许。 

为了提高 /v1/authz 接口的性能，iam-authz-server 将密钥和策略信息缓存在内存中，以便实现快速查询。那密钥和策略信息是如何实现缓存的呢？

首先，iam-authz-server 通过调用 iam-apiserver 提供的 gRPC 接口，将密钥和授权策略信息缓存到内存中。

同时，为了使内存中的缓存信息和 iam-apiserver 中的信息保持一 致，当 iam-apiserver 中有密钥或策略被更新时，iam-apiserver 会往特定的 Redis Channel（iam-authz-server 也会订阅该 Channel）中发送 PolicyChanged 和 SecretChanged 消息。

这样一来，当 iam-authz-server 监听到有新消息时就会获取并解析消息，根据消息内容判断是否需要重新调用 gRPC 接来获取密钥和授权策略信息，再更新到内存中。

#### 第 3 步，授权日志数据分析

iam-authz-server 会将授权日志上报到 Redis 高速缓存中，然后 iam-pump 组件会异步消费这些授权日志，再把清理后的数据保存在 MongoDB 中，供运营系统 iamoperating-system 查询。 

这里还有一点要注意：iam-authz-server 将授权日志保存在 Redis 高性能 key-value 数据库中，可以最大化减少写入延时。不保存在内存中是因为授权日志量没法预测， 当授权日志量很大时，很可能会将内存耗尽，造成服务中断。

#### 第 4 步，运营平台授权数据展示

iam-operating-system 是 IAM 的运营系统，它可以通过查询 MongoDB 获取并展示运营数据，比如某个用户的授权 / 失败次数、授权失败时的授权信息等。

此外，也可以 通过 iam-operating-system 调用 iam-apiserver 服务来做些运营管理工作。比如，以上帝视角查看某个用户的授权策略供排障使用，或者调整用户可创建密钥的最大个数，再或者通过白名单的方式，让某个用户不受密钥个数限制的影响等等。



### IAM 软件架构模式 

在设计软件时，首先要做的就是选择一种软件架构模式，它对软件后续的开发方式、 软件维护成本都有比较大的影响。因此，这里会简单聊聊 2 种最常用的软件架构模式，分别是前后端分离架构和 MVC 架构。 

#### 前后端分离架构

因为 IAM 系统采用的就是前后端分离的架构，所以就以 IAM 的运营系统 iamoperating-system 为例来详细说说这个架构。

一般来说，运营系统的功能可多可少，对于一些具有复杂功能的运营系统，可以采用前后端分离的架构。其中，前端负责页面的展示以及数据的加载和渲染，后端只负责返回前端需要的数据。 

iam-operating-system 前后端分离架构如下图所示。

![image-20211103230514402](IAM-document.assets/image-20211103230514402.png)

采用了前后端分离架构之后，当通过浏览器请求前端 ops-webconsole 时，ops-webconsole 会先请求静态文件服务器加载静态文件，比如 HTML、CSS 和 JavaScript， 然后它会执行 JavaScript，通过负载均衡请求后端数据，最后把后端返回的数据渲染到前端页面中。 

采用前后端分离的架构，让前后端通过 RESTful API 通信，会带来以下 5 点好处：

- 可以让前、后端人员各自专注在自己业务的功能开发上，让专业的人做专业的事，来提高代码质量和开发效率。
- 前后端可并行开发和发布，这也能提高开发和发布效率，加快产品迭代速度。
- 前后端组件、代码分开，职责分明，可以增加代码的维护性和可读性，减少代码改动引起的 Bug 概率，同时也能快速定位 Bug
- 前端 JavaScript 可以处理后台的数据，减少对后台服务器的压力。
- 可根据需要选择性水平扩容前端或者后端来节约成本。

#### MVC 架构 

但是，如果运营系统功能比较少，采用前后端分离框架的弊反而大于利，比如前后端分离要同时维护 2 个组件会导致部署更复杂，并且前后端分离将人员也分开了，这会增加一定程度的沟通成本。

同时，因为代码中也需要实现前后端交互的逻辑，所以会引入一定的开发量。 这个时候，可以尝试直接采用 MVC 软件架构，MVC 架构如下图所示。

![image-20211103230916201](IAM-document.assets/image-20211103230916201.png)

MVC 的全名是 Model View Controller，它是一种架构模式，分为 Model、View、 Controller 三层，每一层的功能如下：

- View（视图）：提供给用户的操作界面，用来处理数据的显示。 
- Controller（控制器）：根据用户从 View 层输入的指令，选取 Model 层中的数据，然后对其进行相应的操作，产生最终结果。 
- Model（模型）：应用程序中用于处理数据逻辑的部分。

MVC 架构的好处是通过控制器层将视图层和模型层分离之后，当更改视图层代码后时，就不需要重新编译控制器层和模型层的代码了。同样，如果业务流程发生改变也只需要变更模型层的代码就可以。

在实际开发中为了更好的 UI 效果，视图层需要经常变更，但是通过 MVC 架构，在变更视图层时，根本不需要对业务逻辑层的代码做任何变化，这 不仅减少了风险还能提高代码变更和发布的效率。 

#### 三层架构

除此之外，还有一种跟 MVC 比较相似的软件开发架构叫三层架构，它包括 UI 层、BLL 层 和 DAL 层。

其中，UI 层表示用户界面，BLL 层表示业务逻辑，DAL 层表示数据访问。

在实际开发中很多人将 MVC 当成三层架构在用，比如说，很多人喜欢把软件的业务逻辑放在 Controller 层里，将数据库访问操作的代码放在 Model 层里，软件最终的代码放在 View 层里，就这样硬生生将 MVC 架构变成了伪三层架构。 

这种代码不仅不伦不类，同时也失去了三层架构和 MVC 架构的核心优势，也就是：通过 Controller 层将 Model 层和 View 层解耦，从而使代码更容易维护和扩展。因此在实际 开发中，也要注意遵循 MVC 架构的开发规范，发挥 MVC 的核心价值。

### 总结

一个好的 Go 应用必须要保证应用的安全性，这可以通过认证和授权来保障。也因此认证和授权是开发一个 Go 项目必须要实现的功能。

为了实现这 2 个功能，并借此机会学习 Go 项目开发，将这 2 个功能升级为一个 IAM 系统。通过讲解如何开发 IAM 系统，来教如何开发 Go 项目。 

要重点掌握 IAM 的功能、架构和使用流程，可以通过 4 步使用流程来了解。 

- 首先，用户通过调用 iam-apiserver 提供的 RESTful API 接口完成注册和登录系统，再调用接口创建密钥和授权策略。
- 创建完密钥对和授权策略之后，IAM 可以通过调用 iam-authz-server 的授权接口完成资源的授权。具体来说，iam-authz-server 通过 gRPC 接口获取 iam-apiserver 中存储的密钥和授权策略信息，通过 JWT 完成认证之后，再通过 ory/ladon 包完成资源的授权。 
- 接着，iam-pump 组件异步消费 Redis 中的数据，并持久化存储在 MongoDB 中，供 iam-operating-system 运营平台展示。
- 最后，IAM 相关的产品、研发人员可以通过 IAM 的运营系统 iam-operating-system 来查看 IAM 系统的使用情况，进行运营分析。例如某个用户的授权 / 失败次数、授权失败时的授权信息等。 

另外，为了提高开发和访问效率，IAM 分别提供了 marmotedu-sdk-go SDK 和 iamctl 命令行工具，二者通过 HTTPS 协议访问 IAM 提供的 RESTful 接口。 



### 课后练习

- 在做 Go 项目开发时经常用到哪些技能点？有些技能点是 IAM 没有包含的？ 
- 在所接触的项目中，哪些是前后端分离架构，哪些是 MVC 架构呢？项目采用的架构是否合理呢？



## 环境准备

想要配置一个 Go 开发环境，可以通过以下 4 步实现：
1. Linux 服务器申请和配置
2. 依赖安装和配置
3. Go 编译环境安装和配置
4. Go 开发 IDE 安装和配置

### Linux 服务器申请和配置

毫无疑问，要安装一个 Go 开发环境，首先需要有一个 Linux 服务器。

Linux 服务器有很多操作系统可供选择，例如：CentOS、Ubuntu、RHEL、Debian 等，但目前生产环境用得最多的还是 CentOS 系统，为了跟生产环境保持一致，选择当前最新的 CentOS版本：CentOS 8.2。

安装一个 Linux 服务器需要两步：服务器申请和配置。

#### Linux 服务器申请

可以通过以下 3 种方式来安装一个 CentOS 8.2 系统。

1. 在物理机上安装一个 CentOS 8.2 系统。
2. 在 Windows/MacBook 上安装虚拟机管理软件，用虚拟机管理软件创建 CentOS 8.2虚拟机。其中，Windows 建议用 VMWare Workstation 来创建虚拟机，MacBook 建议用 VirtualBox 来创建虚拟机。
3. 在诸如腾讯云、阿里云、华为云等平台上购买一个虚拟机，并预装 CentOS 8.2 系统。

#### Linux 服务器配置

申请完 Linux 服务之后，需要通过 SecureCRT 或 Xshell 等工具登录 Linux 服务器，并对服务器做一些简单必要的配置，包括创建普通用户、添加 sudoers、配置 `$HOME/.bashrc` 文件。

第一步，用 Root 用户登录 Linux 系统，并创建普通用户。

一般来说，一个项目会由多个开发人员协作完成，为了节省企业成本，公司不会给每个开发人员都配备一台服务器，而是让所有开发人员共用一个开发机，通过普通用户登录开发机进行开发。

因此，为了模拟真实的企业开发环境，通过一个普通用户的身份来进行项目的开发，创建方法如下：

```bash
$ useradd going  # 创建 going 用户，通过 going 用户登录开发机进行开发
$ passwd going  # 设置密码
Changing password for user going.
New password:
Retype new password:
passwd: all authentication tokens updated successfully.

# ps：如果是刚申请的云服务新手，需要先在云服务的控制台中设置root用户的服务器密码，才可以使用 ssh 进行连接 
```

不仅如此，使用普通用户登录和操作开发机也可以保证系统的安全性，这是一个比较好的习惯，所以在日常开发中也要尽量避免使用 Root 用户。

第二步，添加 sudoers。

很多时候，普通用户也要用到 Root 的一些权限，但 Root 用户的密码一般是由系统管理员维护并定期更改的，每次都向管理员询问密码又很麻烦。因此，将普通用户加入到 sudoers 中，这样普通用户就可以通过 sudo 命令来暂时获取 Root 的权限。

具体来说，可以执行如下命令添加：

```bash
$ sed -i '/^root.*ALL=(ALL).*ALL/a\going\tALL=(ALL) \tALL' /etc/sudoers
```

第三步，用新的用户名（going）和密码（going）登录 Linux 服务器。

这一步也可以验证普通用户是否创建成功。

第四步，配置 `$HOME/.bashrc` 文件。

登录新服务器后的第一步就是配置 `$HOME/.bashrc` 文件，以使 Linux 登录 shell 更加易用，例如配置 LANG 解决中文乱码，配置 PS1 可以避免整行都是文件路径，并将`$HOME/bin` 加入到 PATH 路径中。配置后的内容如下：

```bash
$ sudo vim .bashrc

# User specific aliases and functions
alias rm='rm -i'
alias cp='cp -i'
alias mv='mv -i'

# Source global definitions
if [ -f /etc/bashrc ]; then
. /etc/bashrc
fi

# User specific environment
# Basic envs
export LANG="en_US.UTF-8"  # 设置系统语言为 en_US.UTF-8，避免终端出现中文乱码
export PS1='[\u@dev \W]\$'  
# 默认的 PS1 设置会展示全部的路径，为了防止过长，这里只展示："用户名@dev 最后的目录名"
export WORKSPACE="$HOME/workspace"  # 设置工作目录
export PATH=$HOME/bin:$PATH  # 将 $HOME/bin 目录加入到 PATH 变量中

# Default entry folder
cd $WORKSPACE # 登录系统，默认进入 workspace 目录
```

有一点需要注意，在 export PATH 时，最好把 `$PATH` 放到最后，因为我们添加到目录中的命令是期望被优先搜索并使用的。

配置完 `$HOME/.bashrc` 后，还需要创建工作目录 workspace。将工作文件统一放在 `$HOME/workspace` 目录中，有几点好处。

- 可以使我们的 $HOME 目录保持整洁，便于以后的文件查找和分类。
- 如果哪一天 / 分区空间不足，可以将整个 workspace 目录 mv 到另一个分区中，并在 / 分区中保留软连接，例如：/home/going/workspace -> /data/workspace/。
- 如果哪天想备份所有的工作文件，可以直接备份 workspace。

具体的操作指令是`$ mkdir -p $HOME/workspace`。配置好 $HOME/.bashrc 文件后，就可以执行 `bash` 命令将配置加载到当前 shell 中了。

至此，就完成了 Linux 开发机环境的申请及初步配置。



### 依赖安装和配置

在 Linux 系统上安装 IAM 系统会依赖一些 RPM 包和工具，有些是直接依赖，有些是间接依赖。为了避免后续的操作出现依赖错误，例如，因为包不存在而导致的编译、命令执行错误等，先统一依赖安装和配置。安装和配置步骤如下。

第一步，安装依赖。

首先，在 CentOS 系统上通过 yum 命令来安装所需工具的依赖，安装命令如下：

```bash
$ sudo yum -y install make autoconf automake cmake perl-CPAN libcurl-devel libtool gcc gcc-c++ glibc-headers zlib-devel git-lfs telnet ctags lrzsz jq expat-devel openssl-devel
```

虽然有些 CentOS 8.2 系统已经默认安装这些依赖了，但是为了确保它们都能被安装，仍然尝试安装一遍。

如果系统提示 Package xxx is already installed.，说明已经安装好了，直接忽略即可。

第二步，安装 Git。

因为安装 IAM 系统、执行 go get 命令、安装 protobuf 工具等都是通过 Git 来操作的，所以接下来还需要安装 Git。

由于低版本的 Git 不支持 --unshallow 参数，而 go get 在安装 Go 包时会用到 git fetch --unshallow 命令，因此要安装一个高版本的 Git，具体的安装方法如下：

```bash
$ cd /tmp
$ wget https://mirrors.edge.kernel.org/pub/software/scm/git/git-2.30.2.tar.gz
$ tar -xvzf git-2.30.2.tar.gz
$ cd git-2.30.2/
$ ./configure
$ make
$ sudo make install
$ git --version # 输出 git 版本号，说明安装成功
git version 2.30.2
```

注意啦，按照上面的步骤安装好之后，要把 Git 的二进制目录添加到 PATH 路径中，不然 Git 可能会因为找不到一些命令而报错。可以通过执行以下命令添加目录：

```bash
tee -a $HOME/.bashrc <<'EOF'
# Configure for git
export PATH=/usr/local/libexec/git-core:$PATH
EOF

# PS: 执行tee命令的时候，小心将原本文件的内容清空了，是在不行就使用 vim 进行文件内容编辑，顺便将所有的配置内容放在一起。
```

第三步，配置 Git。

直接执行如下命令配置 Git：

```bash
$ git config --global user.name "rmliu" # 用户名改成自己的
$ git config --global user.email "xxx@gmail.com" # 邮箱改成自己的
$ git config --global credential.helper store # 设置 Git，保存用户名和密码
$ git config --global core.longpaths true  # 解决 Git 中 'Filename too long' 的错误
```

除了按照上述步骤配置 Git 之外，还有几点需要注意。

首先，在 Git 中，会把非 ASCII 字符叫做 Unusual 字符。这类字符在 Git 输出到终端的时候默认是用 8 进制转义字符输出的（以防乱码），但现在的终端多数都支持直接显示非 ASCII 字符，所以可以关闭掉这个特性，具体的命令如下：

```bash
$ git config --global core.quotepath off
```

其次，如果觉得访问 github.com 太慢，可以通过国内 GitHub 镜像网站来访问，配置方法如下：

```bash
$ git config --global url."https://github.com.cnpmjs.org/".insteadOf "https://github.com/"
```

这里要注意，通过镜像网站访问仅对 HTTPS 协议生效，对 SSH 协议不生效，并且 github.com.cnpmjs.org 的同步时间间隔为 1 天。

最后，GitHub 限制最大只能克隆 100M 的仓库，为了能够克隆容量大于 100M 的仓库，还需要安装 Git Large File Storage，安装方式如下：

```bash
$ git lfs install --skip-repo
```

现在就完成了依赖的安装和配置。



### Go 编译环境安装配置

Go 是一门编译型语言，所以在部署 IAM 系统之前，需要将代码编译成可执行的二进制文件。因此需要安装 Go 编译环境。

除了 Go，也会用 gRPC 框架展示 RPC 通信协议的用法，所以也需要将 ProtoBuf 的.proto 文件编译成 Go 语言的接口。因此，也需要安装 ProtoBuf 的编译环境。

#### Go 编译环境安装和配置

安装 Go 语言相对来说比较简单，只需要下载源码包、设置相应的环境变量即可。

首先，从 Go 语言官方网站下载对应的 Go 安装包以及源码包，这里下载的是 go1.16.2 版本：

```bash
$ wget https://golang.org/dl/go1.16.2.linux-amd64.tar.gz -O /tmp/go1.16.2.linux-amd64.tar.gz
```

在下载的时候，要选择：linux-amd64 的格式。

如果因为被墙的原因访问不了 golang.org，也可以执行下面的命令下载 ：

```bash
$ wget https://marmotedu-1254073058.cos.ap-beijing.myqcloud.com/tools/go1.16.2.linux-amd64.tar.gz -O /tmp/go1.16.2.linux-amd64.tar.gz
```

接着，完成解压和安装，命令如下：

```bash
$ mkdir -p $HOME/go
$ tar -xvzf /tmp/go1.16.2.linux-amd64.tar.gz -C $HOME/go
$ mv $HOME/go/go $HOME/go/go1.16.2
```

最后，执行以下命令，将下列环境变量追加到`$HOME/.bashrc`文件中。

```bash
tee -a $HOME/.bashrc <<'EOF'
# Go envs
export GOVERSION=go1.16.2  # Go 版本设置
export GO_INSTALL_DIR=$HOME/go  # Go 安装目录
export GOROOT=$GO_INSTALL_DIR/$GOVERSION  # GOROOT 设置
export GOPATH=$WORKSPACE/golang  # GOPATH 设置
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH  # 将 GO 语言自带的和通过 go install 安装的二进制文件加入到 PATH 路径中
export GO111MODULE="on"  # 开启 Go moudles 特性
export GOPROXY=https://goproxy.cn,direct  # 安装 Go 模块时，代理服务器设置
export GOPRIVATE=
export GOSUMDB=off  # 关闭校验 Go 依赖包的哈希值
EOF
```

为什么要增加这么多的环境变量呢？

这是因为，Go语言就是通过一系列的环境变量来控制Go编译器行为的。因此我们一定要理解每一个环境变量的含义。

![image-20211031152947176](IAM-document.assets/image-20211031152947176.png)

因为 Go 以后会用 Go modules 来管理依赖，所以建议将 GO111MODULE 设置为 on。

在使用模块的时候，` $GOPATH` 是无意义的，不过它还是会把下载的依赖储存在 `$GOPATH/pkg/mod`目录中，也会把 go install 的二进制文件存放在 `$GOPATH/bin`目录中。

还要将 `$GOPATH/bin`、`$GOROOT/bin`加入到Linux可执行文件搜索路径中。这样，就可以直接在 bash shell 中执行 go 自带的命令，以及通过 go install 安装的命令。

最后进行测试，执行 go version 命令，可以成功输出 Go 的版本，就说明 Go 编译环境安装成功。具体命令如下：

```bash
$ bash  # 由于更改了 .bashrc 所以需要重新应用一下新文件
$ go version
go version go1.16.2 linux/amd64
```



#### ProtoBuf 编译环境安装

再来安装 protobuf 的编译器 protoc。

protoc 需要 protoc-gen-go 来完成 Go 语言的代码转换，因此需要安装 protoc 和 protoc-gen-go 这 2 个工具。它们的安装方法比较简单，下面给出代码和操作注释。

```bash
# 第一步：安装 protobuf
$ cd /tmp/
$ git clone --depth=1 https://github.com/protocolbuffers/protobuf
$ cd protobuf
$ ./autogen.sh
$ ./configure
$ make
$ sudo make install
$ protoc --version # 查看 protoc 版本，成功输出版本号，说明安装成功
libprotoc 3.19.0

# 第二步：安装 protoc-gen-go
$ go get -u github.com/golang/protobuf/protoc-gen-go
```

注意：当第一次执行 go set 命令的时候，因为本地无缓存，所以需要下载所有的依赖模块。因为安装速度会比较慢，耐心等待！



### Go 开发 IDE 安装和配置

编译环境准备完之后，还需要一个代码编辑器才能开始 Go 项目开发，并且为了提高开发效率，需要将这个编辑器配置成 Go IDE。

目前，GoLand、VSCode 这些 IDE 都很优秀，使用的也很多，但它们都是 Windows系统下的 IDE。因此，在 Linux 环境下可以选择将 Vim 配置成 Go IDE，熟悉 Vim IDE 的操作之后，它的开发效率不输 GoLand 和 VSCode。

比如说，可以通过 **SpaceVim** 将 Vim 配置成一个 Go IDE。

SpaceVim 是一个社区驱动的模块化的 Vim IDE，以模块的方式组织管理插件以及相关配置， 为不同的语言开发量身定制了相关的开发模块，该模块提供代码自动补全、 语法检查、格式化、调试、REPL等特性。只需要载入相关语言的模块就能得到一个开箱即用的 Vim IDE 了。

Vim 可以选择 **NeoVim**。

NeoVim 是基于 Vim 的一个 fork 分支，它主要解决了 Vim8 之前版本中的异步执行、开发模式等问题，对 Vim 的兼容性很好。同时对 vim 的代码进行了大量地清理和重构，去掉了对老旧系统的支持，添加了新的特性。

虽然 **Vim8** 后来也新增了异步执行等特性，在使用层面两者差异不大，但是 NeoVim 开发更激进，新特性更多，架构也相对更合理，所以选择了 NeoVim，也可以根据个人爱好来选择（都是很优秀的编辑器，这里不做优缺点比较）。

Vim IDE 的安装和配置主要分五步。

第一步，安装 NeoVim。

直接执行 pip3 和 yum 命令安装即可，安装方法如下：

```bash
$ sudo pip3 install pynvim
$ sudo yum -y install neovim
```

第二步，配置 `$HOME/.bashrc`。

先配置 nvim 的别名为 vi，这样当我们执行 vi 时，Linux 系统就会默认调用 nvim。同时，配置 EDITOR 环境变量，可以使一些工具，例如 Git 默认使用 nvim。配置方法如下：

```bash
tee -a $HOME/.bashrc <<'EOF'
# Configure for nvim
export EDITOR=nvim  # 默认的编辑器（git 会用到）
alias vi="nvim"
EOF
```

第三步，检查 nvim 是否安装成功。

可以通过查看 NeoVim 版本来确认是否成功安装，如果成功输出版本号，说明 NeoVim 安装成功。

```bash
$ bash
$ vi --version # 输出 NVIM v0.3.8 说明安装成功
NVIM v0.3.8
Build type: RelWithDebInfo
...
```

第四步，离线安装 SpaceVim。

安装 SpaceVim 步骤稍微有点复杂，为了简化安装，同时消除网络的影响，将安装和配置 SpaceVim 的步骤做成了一个离线安装包
marmotVim 。marmotVim 可以进行 SpaceVim 的安装、卸载、打包等操作，安装步骤如下：

```bash
$ cd /tmp
$ wget https://marmotedu-1254073058.cos.ap-beijing.myqcloud.com/tools/marmotVim.tar.gz
$ tar -xvzf marmotVim.tar.gz
$ cd marmotVim
$ ./marmotVimCtl install
```

SpaceVim 配置文件为：`$HOME/.SpaceVim.d/init.toml` 和`$HOME/.SpaceVim.d/autoload/custom_init.vim`，可自行配置（配置文件中有配置说明）：

- init.toml：SpaceVim 的配置文件
- custom_init.vim：兼容 vimrc，用户自定义的配置文件

SpaceVim Go IDE 常用操作的按键映射如下表所示：

![image-20211031163355137](IAM-document.assets/image-20211031163355137.png)

第五步，Go 工具安装。

SpaceVim 会用到一些 Go 工具，比如在函数跳转时会用到 guru、godef 工具，在格式化时会用到 goimports，所以需要安装这些工具。安装方法有 2 种：

1. Vim 底线命令安装：vi test.go，然后执行：`:GoInstallBinaries` 安装。
2. 拷贝工具：直接将整理好的工具文件拷贝到 `$GOPATH/bin` 目录下。

为了方便，可以直接拷贝我已经打包好的 Go 工具到指定目录下：

```bash
$ cd /tmp
$ wget https://marmotedu-1254073058.cos.ap-beijing.myqcloud.com/tools/gotools-for-spacevim.tgz
$ mkdir -p $GOPATH/bin
$ tar -xvzf gotools-for-spacevim.tgz -C $GOPATH/bin
```



### 总结

这一讲，一起安装和配置了一个 Go 开发环境，为了方便回顾，将安装和配置过程绘制成了一个流程图，如下所示。

![image-20211031164240040](IAM-document.assets/image-20211031164240040.png)

有了这个开发环境，接下来就可以在学习的过程中随时进行编码，来熟悉和验证知识点了，所以一定要先完成这一讲的部署。



### 课后练习

1. 试着编写一个 main.go，在 main 函数中打印 Hello World，并执行 go run main.go 运行代码，测试 Go 开发环境。
2. 试着编写一个 main.go，代码如下：

```go
package main

import "fmt"

func main() {
   fmt.Println("hello world!")
}
```

将鼠标放在 Println 上，键入 Enter 键跳转到函数定义处，键入 Ctrl + I 返回到跳转点。



## 项目部署

部署过程分成 2 大步。

1. 安装和配置数据库：需要安装和配置 MariaDB、Redis 和 MongoDB。
2. 安装和配置 IAM 服务：需要安装和配置 iam-apiserver、iam-authz-server、iam-pump、iamctl 和 man 文件。

### 下载 IAM 项目代码

因为 IAM 的安装脚本存放在 iam 代码仓库中，安装需要的二进制文件也需要通过 iam 代码构建，所以在安装之前，需要先下载 iam 代码：

```bash
$ mkdir -p $WORKSPACE/golang/src/github.com/marmotedu
$ cd $WORKSPACE/golang/src/github.com/marmotedu
$ git clone --depth=1 https://github.com/marmotedu/iam
```

其中，marmotedu 和 marmotedu/iam 目录存放了本实战项目的代码，在学习过程中，需要频繁访问这 2 个目录，为了访问方便，可以追加如下 2 个环境变量和 2 个alias 到$HOME/.bashrc 文件中：

```bash
$ tee -a $HOME/.bashrc << 'EOF'
# Alias for qiuick access
export GOWORK="$WORKSPACE/golang/src"
export IAM_ROOT="$GOWORK/github.com/marmotedu/iam"
alias mm="cd $GOWORK/github.com/marmotedu"
alias i="cd $GOWORK/github.com/marmotedu/iam"
EOF
$ bash
```

之后，可以先通过执行 alias 命令 mm 访问 `$GOWORK/github.com/marmotedu` 目
录，再通过执行 alias 命令 i 访问 `$GOWORK/github.com/marmotedu/iam` 目录。

这里建议善用 alias，将常用操作配置成 alias，方便以后操作。

在安装配置之前需要执行以下命令 export going 用户的密码，这里假设密码是iam59!z$：

```bash
export LINUX_PASSWORD='iam59!z$'

# export LINUX_PASSWORD='going'
# ps: 如果执行export 之后，在安装的过程中，执行 xxx.sh 文件，就可以内部直接使用该环境变量
```



### 安装和配置数据库

因为 IAM 系统用到了 MariaDB、Redis、MongoDB 数据库来存储数据，而 IAM 服务在启动时会先尝试连接这些数据库，所以为了避免启动时连接数据库失败，这里先来安装需要的数据库。

#### 安装和配置 MariaDB

IAM 会把 REST 资源的定义信息存储在关系型数据库中，关系型数据库选择了 MariaDB。

为啥选择 MariaDB，而不是 MySQL 呢？

选择 MariaDB 一方面是因为它是发展最快的 MySQL 分支，相比 MySQL，它加入了很多新的特性，并且它能够完全兼容MySQL，包括 API 和命令行。另一方面是因为 MariaDB 是开源的，而且迭代速度很快。

首先，可以通过以下命令安装和配置 MariaDB，并将 Root 密码设置为 iam59!z$：

```bash
$ cd $IAM_ROOT
$ ./scripts/install/mariadb.sh iam::mariadb::install
```

然后，可以通过以下命令，来测试 MariaDB 是否安装成功。

```bash
$ mysql -h127.0.0.1 -uroot -p'iam59!z$'
MariaDB [(none)]>
```



#### 安装和配置 Redis

在 IAM 系统中，由于 iam-authz-server 是从 iam-apiserver 拉取并缓存用户的密钥 / 策略信息的，因此同一份密钥 / 策略数据会分别存在 2 个服务中，这可能会出现数据不一致的情况。

数据不一致会带来一些问题，例如当通过 iam-apiserver 创建了一对密钥，但是这对密钥还没有被 iam-authz-server 缓存，这时候通过这对密钥访问 iam-authz-server 就会访问失败。

为了保证数据的一致性，可以使用 Redis 的发布订阅 (pub/sub) 功能进行消息通知。

同时，iam-authz-server 也会将授权审计日志缓存到 Redis 中，所以也需要安装 Redis key-value 数据库。

可以通过以下命令来安装和配置 Redis，并将 Redis 的初始密码设置为 iam59!z$ ：

```bash
$ cd $IAM_ROOT
$ ./scripts/install/redis.sh iam::redis::install
```

这里要注意，scripts/install/redis.sh 脚本中 iam::redis::install 函数对 Redis 做了一些配置，例如修改 Redis 使其以守护进程的方式运行、修改 Redis 的密码为 iam59!z$ 等，详细配置可参考函数 iam::redis::install 函数。

安装完成后，可以通过以下命令，来测试 Redis 是否安装成功：

```bash
$ redis-cli -h 127.0.0.1 -p 6379 -a 'iam59!z$'  # 连接 Redis，-h 指定主机，-p 指定监听端口，-a 指定登录密码
127.0.0.1:6379>
```



#### 安装和配置 MongoDB

因为 iam-pump 会将 iam-authz-server 产生的数据处理后存储在 MongoDB 中，所以也需要安装 MongoDB 数据库。

主要分两步安装：首先安装 MongoDB，然后再创建 MongoDB 账号。

##### 第 1 步，安装 MongoDB

首先，可以通过以下 4 步来安装 MongoDB。

1. 配置 MongoDB yum 源，并安装 MongoDB。

CentOS 8.x 系统默认没有配置安装 MongoDB 需要的 yum 源，所以需要先配置好 yum 源再安装：

```bash
$ sudo tee /etc/yum.repos.d/mongodb-org-4.4.repo<<'EOF'
[mongodb-org-4.4]
name=MongoDB Repository
baseurl=https://repo.mongodb.org/yum/redhat/$releasever/mongodb-org/4.4/x86_64
gpgcheck=1
enabled=1
gpgkey=https://www.mongodb.org/static/pgp/server-4.4.asc
EOF

$ sudo yum install -y mongodb-org

# ps: 由于 mongodb 的 yum 源的配置文件不存在，需要进行新建。
```

2. 关闭 SELinux。

在安装的过程中，SELinux 有可能会阻止 MongoDB 访问 /sys/fs/cgroup，所以还需要关闭 SELinux。

```bash
$ sudo setenforce 0
$ sudo sed -i 's/^SELINUX=.*$/SELINUX=disabled/' /etc/selinux/config  # 永久关闭 SELINUX
```

3. 开启外网访问权限和登录验证.

MongoDB 安装完之后，默认情况下是不会开启外网访问权限和登录验证，为了方便使用，建议先开启这些功能，执行如下命令开启：

```bash
$ sudo sed -i '/bindIp/{s/127.0.0.1/0.0.0.0/}' /etc/mongod.conf
$ sudo sed -i '/^#security/a\security:\n authorization: enabled' /etc/mongod.conf
```

4. 启动 MongoDB。

配置完 MongoDB 之后，就可以启动它了，具体的命令如下：

```bash
$ sudo systemctl start mongod
$ sudo systemctl enable mongod  # 设置开机启动
$ sudo systemctl status mongod  # 查看 mongod 运行状态，如果输出中包含 active (running) 字样说明 mongod 成功启动
```

安装完 MongoDB 后，就可以通过 mongo 命令登录 MongoDB Shell。如果没有报错，就说明 MongoDB 被成功安装了。

```bash
$ mongo --quiet "mongodb://127.0.0.1:27017"
>
```

##### 第 2 步，创建 MongoDB 账号

安装完 MongoDB 之后，默认是没有用户账号的，为了方便 IAM 服务使用，需要先创建好管理员账号，通过管理员账户登录 MongoDB，可以执行创建普通用户、数据库等操作。

1. 创建管理员账户。

首先，通过 use admin 指令切换到 admin 数据库，再通过 db.auth("用户名"，"用户密码") 验证用户登录权限。如果返回 1 表示验证成功；如果返回 0 表示验证失败。具体的命令如下：

```bash
$ mongo --quiet "mongodb://127.0.0.1:27017"
> use admin
switched to db admin
> db.createUser({user:"root",pwd:"iam59!z$",roles:["root"]})
Successfully added user: { "user" : "root", "roles" : [ "root" ] }
> db.auth("root", "iam59!z$")
1
```

此外，如果想删除用户，可以使用 db.dropUser("用户名") 命令。

db.createUser 用到了以下 3 个参数。

- user: 用户名。
- pwd: 用户密码。
- roles: 用来设置用户的权限，比如读、读写、写等。

因为 admin 用户具有 MongoDB 的 Root 权限，权限过大安全性会降低。为了提高安全性，还需要创建一个 iam 普通用户来连接和操作 MongoDB。

2. 创建 iam 用户，命令如下：

```bash
$ mongo --quiet mongodb://root:'iam59!z$'@127.0.0.1:27017/tyk_analytics?authSource=admin # 用管理员账户连接 MongoDB
> use iam_analytics
switched to db iam_analytics
> db.createUser({user:"iam",pwd:"iam59!z$",roles:["dbOwner"]})
Successfully added user: { "user" : "iam", "roles" : [ "dbOwner" ] }
> db.auth("iam", "iam59!z$")
1
```

创建完 iam 普通用户后，就可以通过 iam 用户登录 MongoDB 了：

```bash
$ mongo --quiet mongodb://iam:'iam59!z$'@127.0.0.1:27017/iam_analytics?authSource=iam_analytics
```

至此，成功安装了 IAM 系统需要的数据库 MariaDB、Redis 和 MongoDB。



### 安装和配置 IAM 系统

要想完成 IAM 系统的安装，还需要安装和配置 iam-apiserver、iam-authz-server、iam-pump 和 iamctl。

#### 准备工作

在开始安装之前，需要先做一些准备工作，主要有 5 步。

1. 初始化 MariaDB 数据库，创建 iam 数据库。
2. 配置 scripts/install/environment.sh。
3. 创建需要的目录。
4. 创建 CA 根证书和密钥。
5. 配置 hosts。

##### 第 1 步，初始化 MariaDB 数据库，创建 iam 数据库

安装完 MariaDB 数据库之后，需要在 MariaDB 数据库中创建 IAM 系统需要的数据库、表和存储过程，以及创建 SQL 语句保存在 IAM 代码仓库中的 configs/iam.sql 文件中。具体的创建步骤如下。

1. 登录数据库并创建 iam 用户。

```bash
$ cd $IAM_ROOT
$ mysql -h127.0.0.1 -P3306 -uroot -p'iam59!z$'  # 连接 MariaDB，-h 指定主机，-P 指定监听端口，-u 指定登录用户，-p 指定登录密码
MariaDB [(none)]> grant all on iam.* TO iam@127.0.0.1 identified by 'iam59!z$';
Query OK, 0 rows affected (0.000 sec)
MariaDB [(none)]> flush privileges;
Query OK, 0 rows affected (0.000 sec)
```

2. 用 iam 用户登录 MariaDB，执行 iam.sql 文件，创建 iam 数据库。

```bash
$ mysql -h127.0.0.1 -P3306 -uiam -p'iam59!z$'
MariaDB [(none)]> source configs/iam.sql;
MariaDB [iam]> show databases;
+--------------------+
| Database           |
+--------------------+
| iam                |
| information_schema |
| test               |
+--------------------+
3 rows in set (0.000 sec)
```

上面的命令会创建 iam 数据库，并创建以下数据库资源。

- 表：
  - user 是用户表，用来存放用户信息；
  - secret 是密钥表，用来存放密钥信息；
  - policy是策略表，用来存放授权策略信息；
  - policy_audit 是策略历史表，被删除的策略会被转存到该表。
- admin 用户：
  - 在 user 表中，我们需要创建一个管理员用户，用户名是 admin，密码是Admin@2021。
- 存储过程：
  - 删除用户时会自动删除该用户所属的密钥和策略信息。

##### 第 2 步，配置 scripts/install/environment.sh

IAM 组件的安装配置都是通过环境变量文件 scripts/install/environment.sh 进行配置的，所以要先配置好 scripts/install/environment.sh 文件。

这里，可以直接使用默认值，提高安装效率。

##### 第 3 步，创建需要的目录

在安装和运行 IAM 系统的时候，需要将配置、二进制文件和数据文件存放到指定的目录。所以需要先创建好这些目录，创建步骤如下。

```bash
$ cd $IAM_ROOT
$ source scripts/install/environment.sh
$ sudo mkdir -p ${IAM_DATA_DIR}/{iam-apiserver,iam-authz-server,iam-pump}  # 创建 Systemd WorkingDirectory 目录
$ sudo mkdir -p ${IAM_INSTALL_DIR}/bin #创建 IAM 系统安装目录
$ sudo mkdir -p ${IAM_CONFIG_DIR}/cert # 创建 IAM 系统配置文件存放目录
$ sudo mkdir -p ${IAM_LOG_DIR} # 创建 IAM 日志文件存放目录
```

##### 第 4 步， 创建 CA 根证书和密钥

为了确保安全，IAM 系统各组件需要使用 x509 证书对通信进行加密和认证。所以，这里需要先创建 CA 证书。CA 根证书是所有组件共享的，只需要创建一个 CA 证书，后续创建的所有证书都由它签名。

可以使用 CloudFlare 的 PKI 工具集 cfssl 来创建所有的证书。

1. 安装 cfssl 工具集。

可以直接安装 cfssl 已经编译好的二进制文件，cfssl 工具集中包含很多工具，这里需要安装 cfssl、cfssljson、cfssl-certinfo，功能如下。

- cfssl：证书签发工具。
- cfssljson：将 cfssl 生成的证书（json 格式）变为文件承载式证书。

这两个工具的安装方法如下：

```bash
$ cd $IAM_ROOT
$ ./scripts/install/install.sh iam::install::install_cfssl

# ps: 执行这个命令时候，需要的安装时间会比较长，30分钟左右，耐心，出错之后，可以选择重新安装，总会安装成功。
```

2. 创建配置文件。

CA 配置文件是用来配置根证书的使用场景 (profile) 和具体参数 (usage、过期时间、服务端认证、客户端认证、加密等)，可以在签名其它证书时用来指定特定场景：

```bash
$ cd $IAM_ROOT
$ tee ca-config.json << EOF
{
   "signing":{
      "default":{
         "expiry":"87600h"
      },
      "profiles":{
         "iam":{
            "usages":[
               "signing",
               "key encipherment",
               "server auth",
               "client auth"
            ],
            "expiry":"876000h"
         }
      }
   }
}
EOF
```

上面的 JSON 配置中，有一些字段解释如下。

- signing：表示该证书可用于签名其它证书（生成的 ca.pem 证书中 CA=TRUE）。
- server auth：表示 client 可以用该证书对 server 提供的证书进行验证。
- client auth：表示 server 可以用该证书对 client 提供的证书进行验证。
- expiry：876000h，证书有效期设置为 100 年。

3. 创建证书签名请求文件。

创建用来生成 CA 证书签名请求（CSR）的 JSON 配置文件：

```bash
$ cd $IAM_ROOT
$ tee ca-csr.json << EOF
{
   "CN":"iam-ca",
   "key":{
      "algo":"rsa",
      "size":2048
   },
   "names":[
      {
         "C":"CN",
         "ST":"BeiJing",
         "L":"BeiJing",
         "O":"marmotedu",
         "OU":"iam"
      }
   ],
   "ca":{
      "expiry":"876000h"
   }
}
EOF
```

上面的 JSON 配置中，有一些字段解释如下。

- C：Country，国家。
- ST：State，省份。
- L：Locality (L) or City，城市。
- CN：Common Name，iam-apiserver 从证书中提取该字段作为请求的用户名 (User Name) ，浏览器使用该字段验证网站是否合法。
- O：Organization，iam-apiserver 从证书中提取该字段作为请求用户所属的组(Group)。
- OU：Company division (or Organization Unit – OU)，部门 / 单位。

除此之外，还有两点需要注意。

- 不同证书 csr 文件的 CN、C、ST、L、O、OU 组合必须不同，否则可能出现 PEER'S CERTIFICATE HAS AN INVALID SIGNATURE 错误。
- 后续创建证书的 csr 文件时，CN、OU 都不相同（C、ST、L、O 相同），以达到区分的目的。

4. 创建 CA 证书和私钥。

首先，通过 cfssl gencert 命令来创建：

```bash
$ cd $IAM_ROOT
$ source scripts/install/environment.sh
$ cfssl gencert -initca ca-csr.json | cfssljson -bare ca
$ ls ca*
ca-config.json ca.csr ca-csr.json ca-key.pem ca.pem
$ sudo mv ca* ${IAM_CONFIG_DIR}/cert # 需要将证书文件拷贝到指定文件夹下（分发证书），方便各组件引用
```

上述命令会创建运行 CA 所必需的文件 ca-key.pem（私钥）和 ca.pem（证书），还会生成 ca.csr（证书签名请求），用于交叉签名或重新签名。

创建完之后，可以通过 cfssl certinfo 命名查看 cert 和 csr 信息：

```bash
$ cfssl certinfo -cert ${IAM_CONFIG_DIR}/cert/ca.pem # 查看 cert(证书信息)
$ cfssl certinfo -csr ${IAM_CONFIG_DIR}/cert/ca.csr # 查看 CSR(证书签名请求)信息
```

##### 第 5 步，配置 hosts

iam 通过域名访问 API 接口，因为这些域名没有注册过，还不能在互联网上解析，所以需要配置 hosts，具体的操作如下：

```bash
$ sudo tee -a /etc/hosts <<EOF
127.0.0.1 iam.api.marmotedu.com
127.0.0.1 iam.authz.marmotedu.com
EOF
```



#### 安装和配置 iam-apiserver

完成了准备工作之后，就可以安装 IAM 系统的各个组件了。首先通过以下 3 步来安装 iam-apiserver 服务。

##### 第 1 步，创建 iam-apiserver 证书和私钥

其它服务为了安全都是通过 HTTPS 协议访问 iam-apiserver，所以要先创建 iam-apiserver 证书和私钥。

1. 创建证书签名请求：

```bash
$ cd $IAM_ROOT
$ source scripts/install/environment.sh
$ tee iam-apiserver-csr.json <<EOF
{
   "CN":"iam-apiserver",
   "key":{
      "algo":"rsa",
      "size":2048
   },
   "names":[
      {
         "C":"CN",
         "ST":"BeiJing",
         "L":"BeiJing",
         "O":"marmotedu",
         "OU":"iam-apiserver"
      }
   ],
   "hosts":[
      "127.0.0.1",
      "localhost",
      "iam.api.marmotedu.com"
   ]
}
EOF
```

代码中的 hosts 字段是用来指定授权使用该证书的 IP 和域名列表，上面的 hosts 列出了 iam-apiserver 服务的 IP 和域名。

2. 生成证书和私钥：

```bash
$ cfssl gencert -ca=${IAM_CONFIG_DIR}/cert/ca.pem \
-ca-key=${IAM_CONFIG_DIR}/cert/ca-key.pem \
-config=${IAM_CONFIG_DIR}/cert/ca-config.json \
-profile=iam iam-apiserver-csr.json | cfssljson -bare iam-apiserver
$ sudo mv iam-apiserver*pem ${IAM_CONFIG_DIR}/cert # 将生成的证书和私钥文件拷贝到配置文件目录
```

##### 第 2 步，安装并运行 iam-apiserver

iam-apiserver 作为 iam 系统的核心组件，需要第一个安装。

1. 安装 iam-apiserver 可执行程序：

```bash
$ cd $IAM_ROOT
$ source scripts/install/environment.sh
$ make build BINS=iam-apiserver
# ps: 如果报错，安装 gin-jwt，go get github.com/appleboy/gin-jwt/v2
$ sudo cp _output/platforms/linux/amd64/iam-apiserver ${IAM_INSTALL_DIR}/bin
```

2. 生成并安装 iam-apiserver 的配置文件（iam-apiserver.yaml）：

```bash
$ ./scripts/genconfig.sh scripts/install/environment.sh configs/iam-apiserver.yaml > iam-apiserver.yaml
$ sudo mv iam-apiserver.yaml ${IAM_CONFIG_DIR}
```

3. 创建并安装 iam-apiserver systemd unit 文件：

```bash
$ ./scripts/genconfig.sh scripts/install/environment.sh init/iam-apiserver.service > iam-apiserver.service
$ sudo mv iam-apiserver.service /etc/systemd/system/
```

4. 启动 iam-apiserver 服务：

```bash
$ sudo systemctl daemon-reload
$ sudo systemctl enable iam-apiserver
$ sudo systemctl restart iam-apiserver
$ systemctl status iam-apiserver # 查看 iam-apiserver 运行状态，如果输出中包含 active (running)字样说明 iam-apiserver 成功启动
```

##### 第 3 步，测试 iam-apiserver 是否成功安装

测试 iam-apiserver 主要是测试 RESTful 资源的 CURD：用户 CURD、密钥 CURD、授权策略 CURD。

首先，需要获取访问 iam-apiserver 的 Token，请求如下 API 访问：

```bash
$ curl -s -XPOST -H'Content-Type: application/json' -d'{"username":"admin","password":"Admin@2021"}'  http://127.0.0.1:8080/login | jq -r .token
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s
```

代码中下面的 HTTP 请求通过-H'Authorization: Bearer <Token>' 指定认证头信息，将上面请求的 Token 替换 <Token> 。

###### 用户 CURD

创建用户、列出用户、获取用户详细信息、修改用户、删除单个用户、批量删除用户，请求方法如下：

```bash
# 创建用户
$ curl -s -XPOST -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' -d'{"password":"User@2021","metadata":{"name":"colin"},"nickname":"colin","email":"colin@foxmail.com","phone":"1812884xxxx"}' http://127.0.0.1:8080/v1/users
# ps: 自己尝试创建用户
curl -s -XPOST -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' -d'{"password":"User@2021","metadata":{"name":"rmliu"},"nickname":"rmliu","email":"rmliu@foxmail.com","phone":"1832884xxxx"}' http://127.0.0.1:8080/v1/users

# 列出用户
$ curl -s -XGET -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/users?offset=0&limit=10

# 获取 colin 用户的详细信息
$ curl -s -XGET -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/users/colin

# 修改 colin 用户
$ curl -s -XPUT -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' -d'{"nickname":"colin","email":"colin_modified@foxmail.com","phone":"1812884xxxx"}' http://127.0.0.1:8080/v1/users/colin

# 删除 colin 用户
$ curl -s -XDELETE -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/users/colin

# 批量删除用户
$ curl -s -XDELETE -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/users?name=colin&name=mark&name=john
```

###### 密钥 CURD

创建密钥、列出密钥、获取密钥详细信息、修改密钥、删除密钥请求方法如下：

```bash
# 创建 secret0 密钥
$ curl -s -XPOST -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' -d'{"metadata":{"name":"secret0"},"expires":0,"description":"admin secret"}' http://127.0.0.1:8080/v1/secrets

# 列出所有密钥
$ curl -s -XGET -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/secrets

# 获取 secret0 密钥的详细信息
$ curl -s -XGET -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/secrets/secret0

# 修改 secret0 密钥
$ curl -s -XPUT -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' -d'{"metadata":{"name":"secret0"},"expires":0,"description":"admin secret(modified)"}' http://127.0.0.1:8080/v1/secrets/secret0

# 删除 secret0 密钥
$ curl -s -XDELETE -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/secrets/secret0
```

这里要注意，因为密钥属于重要资源，被删除会导致所有的访问请求失败，所以密钥不支持批量删除。

###### 授权策略 CURD

创建策略、列出策略、获取策略详细信息、修改策略、删除策略请求方法如下：

```bash
# 创建策略
$ curl -s -XPOST -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s'-d'{"metadata":{"name":"policy0"},"policy":{"description":"One policy to rule them all.","subjects":["users:<peter|ken>","users:maria","groups:admins"],"actions":["delete","<create|update>"],"effect":"allow","resources":["resources:articles:<.*>","resources:printer"],"conditions":{"remoteIP":{"type":"CIDRCondition","options":{"cidr":"192.168.0.1/16"}}}}}' http://127.0.0.1:8080/v1/policies

# 列出所有策略
$ curl -s -XGET -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/policies

# 获取 policy0 策略的详细信息
$ curl -s -XGET -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/policies/policy0

# 修改 policy 策略
$ curl -s -XPUT -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' -d'{"metadata":{"name":"policy0"},"policy":{"description":"One policy to rule them all(modified).","subjects":["users:<peter|ken>","users:maria","groups:admins"],"actions":["delete","<create|update>"],"effect":"allow","resources":["resources:articles:<.*>","resources:printer"],"conditions":{"remoteIP":{"type":"CIDRCondition","options":{"cidr":"192.168.0.1/16"}}}}}' http://127.0.0.1:8080/v1/policies/policy0

# 删除 policy0 策略
$ curl -s -XDELETE -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' http://127.0.0.1:8080/v1/policies/policy0
```

#### 安装 iamctl

上面，安装了 iam 系统的 API 服务。但是想要访问 iam 服务，还需要安装客户端工具 iamctl。具体来说，可以通过 3 步完成 iamctl 的安装和配置。

##### 第 1 步，创建 iamctl 证书和私钥

iamctl 使用 https 协议与 iam-apiserver 进行安全通信，iam-apiserver 对 iamctl 请求包含的证书进行认证和授权。iamctl 后续用于 iam 系统访问和管理，所以这里创建具有最高权限的 admin 证书。

1. 创建证书签名请求。

下面创建的证书只会被 iamctl 当作 client 证书使用，所以 hosts 字段为空。代码如下：

```bash
$ cd $IAM_ROOT
$ source scripts/install/environment.sh
$ cat > admin-csr.json <<EOF
{
   "CN":"admin",
   "key":{
      "algo":"rsa",
      "size":2048
   },
   "names":[
      {
         "C":"CN",
         "ST":"BeiJing",
         "L":"BeiJing",
         "O":"marmotedu",
         "OU":"iamctl"
      }
   ],
   "hosts":[
      
   ]
}
EOF
```

2. 生成证书和私钥：

```bash
$ cfssl gencert -ca=${IAM_CONFIG_DIR}/cert/ca.pem \
-ca-key=${IAM_CONFIG_DIR}/cert/ca-key.pem \
-config=${IAM_CONFIG_DIR}/cert/ca-config.json \
-profile=iam admin-csr.json | cfssljson -bare admin
$ mkdir -p $(dirname ${CONFIG_USER_CLIENT_CERTIFICATE}) $(dirname ${CONFIG_USER_CLIENT_KEY}) # 创建客户端证书存放的目录
$ mv admin.pem ${CONFIG_USER_CLIENT_CERTIFICATE} # 安装 TLS 的客户端证书
$ mv admin-key.pem ${CONFIG_USER_CLIENT_KEY} # 安装 TLS 的客户端私钥文件
```

##### 第 2 步，安装 iamctl

iamctl 是 IAM 系统的客户端工具，其安装位置和 iam-apiserver、iam-authz-server、iam-pump 位置不同，为了能够在 shell 下直接运行 iamctl 命令，需要将 iamctl 安装到`$HOME/bin` 下，同时将 iamctl 的配置存放在默认加载的目录下：`$HOME/.iam`。主要分 2 步进行。

1. 安装 iamctl 可执行程序：

```bash
$ cd $IAM_ROOT
$ source scripts/install/environment.sh
$ make build BINS=iamctl
$ cp _output/platforms/linux/amd64/iamctl $HOME/bin
```

2. 生成并安装 iamctl 的配置文件（config）：

```bash
$ ./scripts/genconfig.sh scripts/install/environment.sh configs/iamctl.yaml > config  # 这是一个 bug
 # 下面的方式亲测可用
./scripts/genconfig.sh scripts/install/environment.sh configs/iamctl.yaml > iamctl.yaml

$ mkdir -p $HOME/.iam
$ mv config $HOME/.iam  # 这是一个 bug
 # 下面的方式亲测可用
mv iamctl.yaml $HOME/.iam
```

因为 iamctl 是一个客户端工具，可能会在多台机器上运行。为了简化部署 iamctl 工具的复杂度，可以把 config 配置文件中跟 CA 认证相关的 CA 文件内容用 base64 加密后，放置在 config 配置文件中。

具体的思路就是把 config 文件中的配置项 client-certificate、client-key、certificate-authority 分别用如下配置项替换 client-certificate-data、client-key-data、certificate-authority-data。这些配置项的值可以通过对 CA 文件使用 base64 加密获得。

假如，certificate-authority 值为/etc/iam/cert/ca.pem，则 certificate-authority-data 的值为 cat "/etc/iam/cert/ca.pem" | base64 | tr -d '\r\n'，其它-data 变量的值类似。这样当再部署 iamctl 工具时，只需要拷贝iamctl 和配置文件，而不用再拷贝 CA 文件了。

##### 第 3 步，测试 iamctl 是否成功安装

执行 iamctl user list 可以列出预创建的 admin 用户，如下图所示：

![image-20211102005041406](IAM-document.assets/image-20211102005041406.png)



#### 安装和配置 iam-authz-server

接下来，需要安装另外一个核心组件：iam-authz-server，可以通过以下 3 步来安装。

##### 第 1 步，创建 iam-authz-server 证书和私钥

1. 创建证书签名请求：

```bash
$ cd $IAM_ROOT
$ source scripts/install/environment.sh
$ tee iam-authz-server-csr.json <<EOF
{
   "CN":"iam-authz-server",
   "key":{
      "algo":"rsa",
      "size":2048
   },
   "names":[
      {
         "C":"CN",
         "ST":"BeiJing",
         "L":"BeiJing",
         "O":"marmotedu",
         "OU":"iam-authz-server"
      }
   ],
   "hosts":[
      "127.0.0.1",
      "localhost",
      "iam.authz.marmotedu.com"
   ]
}
EOF
```

代码中的 hosts 字段指定授权使用该证书的 IP 和域名列表，上面的 hosts 列出了 iam-authz-server 服务的 IP 和域名。

2. 生成证书和私钥：

```bash
$ cfssl gencert -ca=${IAM_CONFIG_DIR}/cert/ca.pem \
-ca-key=${IAM_CONFIG_DIR}/cert/ca-key.pem \
-config=${IAM_CONFIG_DIR}/cert/ca-config.json \
-profile=iam iam-authz-server-csr.json | cfssljson -bare iam-authz-server
$ sudo mv iam-authz-server*pem ${IAM_CONFIG_DIR}/cert # 将生成的证书和私钥文件拷贝到配置文件目录
```

##### 第 2 步，安装并运行 iam-authz-server

安装 iam-authz-server 步骤和安装 iam-apiserver 步骤基本一样，也需要 4 步。

1. 安装 iam-authz-server 可执行程序：

```bash
$ cd $IAM_ROOT
$ source scripts/install/environment.sh
$ make build BINS=iam-authz-server
$ sudo cp _output/platforms/linux/amd64/iam-authz-server ${IAM_INSTALL_DIR}/bin
```

2. 生成并安装 iam-authz-server 的配置文件（iam-authz-server.yaml）：

```bash
$ ./scripts/genconfig.sh scripts/install/environment.sh configs/iam-authz-server.yaml > iam-authz-server.yaml
$ sudo mv iam-authz-server.yaml ${IAM_CONFIG_DIR}
```

3. 创建并安装 iam-authz-server systemd unit 文件：

```bash
$ ./scripts/genconfig.sh scripts/install/environment.sh init/iam-authz-server.service > iam-authz-server.service
$ sudo mv iam-authz-server.service /etc/systemd/system/
```

4. 启动 iam-authz-server 服务：

```bash
$ sudo systemctl daemon-reload
$ sudo systemctl enable iam-authz-server
$ sudo systemctl restart iam-authz-server
$ systemctl status iam-authz-server # 查看 iam-authz-server 运行状态，如果输出中包含 active (running)字样说明 iam-authz-server 成功启动。
```

##### 第 3 步，测试 iam-authz-server 是否成功安装

1. 创建授权策略。

```bash
$ curl -s -XPOST -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' -d '{"metadata":{"name":"authztest"},"policy":{"description":"One policy to rule them all.","subjects":["users:<peter|ken>","users:maria","groups:admins"],"actions":["delete","<create|update>"],"effect":"allow","resources":["resources:articles:<.*>","resources:printer"],"conditions":{"remoteIP":{"type":"CIDRCondition","options":{"cidr":"192.168.0.1/16"}}}}}' http://127.0.0.1:8080/v1/policies
```

2. 创建密钥，并从代码的输出中提取 secretID 和 secretKey。

```bash
$ curl -s -XPOST -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzU5NTM1ODcsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzU4NjcxODcsInN1YiI6ImFkbWluIn0.b7SyEqpCLuR_LnyJpA3IjFkUiQ-LPHBq3QXMjEqQZ4s' -d'{"metadata":{"name":"authztest"},"expires":0,"description":"admin secret"}' http://127.0.0.1:8080/v1/secrets

# 输出
{"metadata":{"id":23,"instanceID":"secret-yj8m30","name":"authztest","createdAt":"2021-11-03T00:55:44.444+08:00","updatedAt":"2021-11-03T00:55:44.444+08:00"},"username":"admin","secretID":"SuXnTvmGOWu5f95BfonhvYi8uxLBH2y6BOlc","secretKey":"6dF1ENyDWBDGlmR6ipUbUcpkdjgqF5Gh","expires":0,"description":"admin secret"}
```

3. 生成访问 iam-authz-server 的 Token。

amctl 提供了 jwt sigin 命令，可以根据 secretID 和 secretKey 签发 Token，方便使用。

```bash
$ iamctl jwt sign SuXnTvmGOWu5f95BfonhvYi8uxLBH2y6BOlc 6dF1ENyDWBDGlmR6ipUbUcpkdjgqF5Gh # iamctl jwt sign $secretID $secretKey

# 输出
eyJhbGciOiJIUzI1NiIsImtpZCI6IlN1WG5Udm1HT1d1NWY5NUJmb25odllpOHV4TEJIMnk2Qk9sYyIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXV0aHoubWFybW90ZWR1LmNvbSIsImV4cCI6MTYzNTg3OTQ3MywiaWF0IjoxNjM1ODcyMjczLCJpc3MiOiJpYW1jdGwiLCJuYmYiOjE2MzU4NzIyNzN9.MXDw-JCkrB4_cT2cBw7il0ss5yTaW3xa3qn_RyFU7P4
```

如果开发过程中有些重复性的操作，为了方便使用，也可以将这些操作以 iamctl 子命令的方式集成到 iamctl 命令行中。

4. 测试资源授权是否通过。

可以通过请求 /v1/authz 来完成资源授权：

```bash
$ curl -s -XPOST -H'Content-Type: application/json' -H'Authorization: Bearer eyJhbGciOiJIUzI1NiIsImtpZCI6IlN1WG5Udm1HT1d1NWY5NUJmb25odllpOHV4TEJIMnk2Qk9sYyIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXV0aHoubWFybW90ZWR1LmNvbSIsImV4cCI6MTYzNTg3OTQ3MywiaWF0IjoxNjM1ODcyMjczLCJpc3MiOiJpYW1jdGwiLCJuYmYiOjE2MzU4NzIyNzN9.MXDw-JCkrB4_cT2cBw7il0ss5yTaW3xa3qn_RyFU7P4' -d'{"subject":"users:maria","action":"delete","resource":"resources:articles:ladon-introduction","context":{"remoteIP":"192.168.0.5"}}' http://127.0.0.1:9090/v1/authz

# 输出
{"allowed":true}
```

如果授权通过会返回：{"allowed":true} 。



#### 安装和配置 iam-pump

安装 iam-pump 步骤和安装 iam-apiserver、iam-authz-server 步骤基本一样，具体步骤如下。

第 1 步，安装 iam-pump 可执行程序。

```bash
$ cd $IAM_ROOT
$ source scripts/install/environment.sh
$ make build BINS=iam-pump
$ sudo cp _output/platforms/linux/amd64/iam-pump ${IAM_INSTALL_DIR}/bin
```

第 2 步，生成并安装 iam-pump 的配置文件（iam-pump.yaml）

```bash
$ ./scripts/genconfig.sh scripts/install/environment.sh configs/iam-pump.yaml > iam-pump.yaml
$ sudo mv iam-pump.yaml ${IAM_CONFIG_DIR}
```

第 3 步，创建并安装 iam-pump systemd unit 文件。

```bash
$ ./scripts/genconfig.sh scripts/install/environment.sh init/iam-pump.service > iam-pump.service
$ sudo mv iam-pump.service /etc/systemd/system/
```

第 4 步，启动 iam-pump 服务。

```bash
$ sudo systemctl daemon-reload
$ sudo systemctl enable iam-pump
$ sudo systemctl restart iam-pump
$ systemctl status iam-pump # 查看 iam-pump 运行状态，如果输出中包含 active (running)字样说明 iam-pump 成功启动。
```

第 5 步，测试 iam-pump 是否成功安装。

```bash
$ curl http://127.0.0.1:7070/healthz

# 输出
{"status": "ok"}
```

经过上面这 5 个步骤，如果返回 {"status": "ok"} 就说明 iam-pump 服务健康。



#### 安装 man 文件

IAM 系统通过组合调用包：github.com/cpuguy83/go-md2man/v2/md2man 和github.com/spf13/cobra 的相关函数生成了各个组件的 man1 文件，主要分 3 步实现。

第 1 步，生成各个组件的 man1 文件。

```bash
$ cd $IAM_ROOT
$ ./scripts/update-generated-docs.sh
```

第 2 步，安装生成的 man1 文件。

```bash
$ sudo cp docs/man/man1/* /usr/share/man/man1/
```

第 3 步，检查是否成功安装 man1 文件。

```bash
$ man iam-apiserver
```

执行 man iam-apiserver 命令后，会弹出 man 文档界面，如下图所示：

![image-20211103011401365](IAM-document.assets/image-20211103011401365.png)

至此，IAM 系统所有组件都已经安装成功了，可以通过 iamctl version 查看客户端和服务端版本，代码如下：

```bash
$ iamctl version -o yaml

# 输出
clientVersion:
  buildDate: "2021-11-02T16:16:11Z"
  compiler: gc
  gitCommit: 5d2922bdca8c65725f4b363e3017595a860602f4
  gitTreeState: dirty
  gitVersion: 5d2922b
  goVersion: go1.16.2
  platform: linux/amd64
serverVersion:
  buildDate: "2021-11-02T15:22:23Z"
  compiler: gc
  gitCommit: 5d2922bdca8c65725f4b363e3017595a860602f4
  gitTreeState: dirty
  gitVersion: 5d2922b
  goVersion: go1.16.2
  platform: linux/amd64
```



### 总结

这一讲，一步一步安装了 IAM 应用，完成安装的同时，也希望能加深对 IAM 应用的理解，并为后面的实战准备好环境。为了更清晰地展示安装流程，把整个安装步骤梳理成了一张脑图。

![image-20211103011821246](IAM-document.assets/image-20211103011821246.png)

所有组件设置的密码都是 iam59!z$，一定要记住啦。



### 课后练习

试着调用 iam-apiserver 提供的 API 接口创建一个用户：xuezhang，并在该用户下创建 policy 和 secret 资源。最后调用 iam-authz-server 提供的/v1/authz 接口进行资源鉴权。



### 一键安装

可以直接执行如下脚本，来完成 IAM 系统的安装：

```bash
$ export LINUX_PASSWORD='iam59!z$' # 重要：这里要 export going 用户的密码
# export LINUX_PASSWORD='going'

$ version=v1.0.0 && curl https://marmotedu-1254073058.cos.ap-beijing.myqcloud.
$ cd /tmp/iam/ && ./scripts/install/install.sh iam::install::install
```



## 规范设计之开源、文档、版本规范

无规矩不成方圆，生活如此，软件开发也是如此。一个应用基本都是多人协作开发的，但不同人的开发习惯、方式都不同。如果没有一个统一的规范，就会造成非常多的问题，比如：

- 代码风格不一：代码仓库中有多种代码风格，读 / 改他人的代码都是一件痛苦的事情， 整个代码库也会看起来很乱。 
- 目录杂乱无章：相同的功能被放在不同的目录，或者一个目录根本不知道它要完成什么功能，新开发的代码也不知道放在哪个目录或文件。这些都会严重降低代码的可维护性。
- 接口不统一：对外提供的 API 接口不统一，例如修改用户接口为 /v1/users/colin， 但是修改密钥接口为 /v1/secret?name=secret0，难以理解和记忆。 
- 错误码不规范：错误码会直接暴露给用户，主要用于展示错误类型，以定位错误问题。 错误码不规范会导致难以辨别错误类型，或者同类错误拥有不同错误码，增加理解难度。

因此，在设计阶段、编码之前，需要一个好的规范来约束开发者，以确保大家开发的是“一个应用”。一个好的规范不仅可以提高软件质量，还可以提高软件的开发效率，降低维护成本，甚至能减少 Bug 数，也可以使开发体验如行云流水一般顺畅。

所以，在编码之前，有必要花一些时间和团队成员一起讨论并制定规范。 

那么，有哪些地方需要制定规范，这些规范又该如何制定呢？ 

### 有哪些地方需要制定规范？

一个 Go 项目会涉及很多方面，所以也会有多种规范，同类规范也会因为团队差异而有所不同。所以，只给讲一些开发中常用的规范。为了便于记忆，根据是否跟代码相关，将它们分为非编码类规范和编码类规范：

- 非编码类规范，主要包括开源规范、文档规范、版本规范、Commit 规范和发布规范。 
- 编码类规范，则主要包括目录规范、代码规范、接口规范、日志规范和错误码规范。

将这些规范整理成了下面一张图：

![image-20211104215950419](IAM-document.assets/image-20211104215950419.png)

先来说说开源规范、文档规范和版本规范，因为 Commit 规范比较多，放到下一讲。至于其他规范，会在后面内容中介绍。例如日志规范，因为和日志设计结合 比较紧密，会放在日志包设计中一起讲。

### 开源规范

首先，来介绍下开源规范。 

其实业界并没有一个官方的开源规范，实际开发中，也很少有人提这个。那么，为什么一定要知道开源规范呢？ 

原因主要有两方面：

- 一是，开源项目在代码质量、代码规范、文档等方面，要比非开源项目要求更高，在项目开发中按照开源项目的要求来规范自己的项目，可以更好地驱动项目质量的提高；
- 二是，一些大公司为了不重复造轮子，会要求公司团队能够将自己的项目开 源，所以提前按开源标准来驱动 Go 项目开发，也会为日后代码开源省去不少麻烦。 

一个开源项目一定需要一个开源协议，开源协议规定了在使用开源软件时的权利和责任，也就是规定了可以做什么，不可以做什么。所以，开源规范的第一条规范就是选择一个合适的开源协议。那么有哪些开源协议，如何选择呢？接下来，来详细介绍下。 

#### 开源协议概述

首先要说明的是，只有开源项目才会用到开源协议，如果项目不准备开源，就用不到开源协议。但先了解一下总是没错的，以后总能用得上。 

业界有上百种开源协议，每种开源协议的要求不一样，有的协议对使用条件要求比较苛刻，有的则相对比较宽松。没必要全都记住，只需要知道经常使用的 6 种开源协议， 也就是 GPL、MPL、LGPL、Apache、BSD 和 MIT 就可以了。至于它们的介绍，可以参考开源协议介绍 。 

那具体如何选择适合自己的开源协议呢？可以参考乌克兰程序员 Paul Bagwell 画的这张图：

![image-20211104220546689](IAM-document.assets/image-20211104220546689.png)

在上图中，右边的协议比左边的协议宽松，在选择时，可以根据菱形框中的选择项从上到下进行选择。为了能够毫无负担地使用 IAM 项目提供的源码，选择了最宽松的 MIT 协议。

另外，因为 Apache 是对商业应用友好的协议，使用者也可以在需要的时候修改代码来满足需要，并作为开源或商业产品发布 / 销售，所以大型公司的开源项目通常会采用 Apache 2.0 开源协议。 

#### 开源规范具有哪些特点？ 

那在参与开源项目，或者按照开源项目的要求来规范代码时，需要关注哪些方面的规范呢？ 其实，一切能让项目变得更优秀的规范，都应该属于开源规范。

开源项目的代码，除了要遵守上面所说的编码类规范和非编码类规范之外，还要遵守下面几个规范。 

- 第一，开源项目，应该有一个高的单元覆盖率。这样，一方面可以确保第三方开发者在开发完代码之后，能够很方便地对整个项目做详细的单元测试，另一方面也能保证提交代码的质量。 
- 第二，要确保整个代码库和提交记录中，不能出现内部 IP、内部域名、密码、密钥这类信息。否则，就会造成敏感信息外漏，可能会对内部业务造成安全隐患。 
- 第三，当开源项目被别的开发者提交 pull request、issue、评论时，要及时处理， 一方面可以确保项目不断被更新，另一方面也可以激发其他开发者贡献代码的积极性。 
- 第四，好的开源项目，应该能够持续地更新功能，修复 Bug。对于一些已经结项、不维护的开源项目，需要及时地对项目进行归档，并在项目描述中加以说明。 

上面这些，是开源规范中比较重要的几点。如果想了解详细的开源规范包括哪些内容，可以看 GitHub 上的这份资料 。 

最后提醒两件事：

- 第一件，如果有条件，可以宣传、运营开源项目，让更多的人知道、使用、贡献代码。比如，可以在掘金、简书等平台发表文章，也可以创建 QQ、微 信交流群等，都是不错的方式。
- 第二件，如果英文好、有时间，文档最好有中英文 2 份，优先使用英文，让来自全球的开发者都能了解、使用和参与项目。

### 文档规范

工作中发现，很多开发者非常注重代码产出，但不注重文档产出。他们觉得，即使没有软件文档也没太大关系，不影响软件交付。这种看法是错误的！因为文档属 于软件交付的一个重要组成部分，没有文档的项目很难理解、部署和使用。

因此，编写文档是一个必不可少的开发工作。那么一个项目需要编写哪些文档，又该如何编写呢？

项目中最需要的 3 类文档是 README 文档、项目文档和 API 接口文档。 下面，一一来说它们的编写规范。

#### README 规范

README 文档是项目的门面，它是开发者学习项目时第一个阅读的文档，会放在项目的根目录下。因为它主要是用来介绍项目的功能、安装、部署和使用的，所以它是可以规范化的。 

下面，直接通过一个 README 模板，来看一下 README 规范中的内容：

```markdown
# 项目名称

<!-- 写一段简短的话描述项目 -->

## 功能特性

<!-- 描述该项目的核心功能点 -->

## 软件架构(可选)

<!-- 可以描述下项目的架构 -->

## 快速开始

### 依赖检查

<!-- 描述该项目的依赖，比如依赖的包、工具或者其他任何依赖项 -->

### 构建

<!-- 描述如何构建该项目 -->

### 运行

<!-- 描述如何运行该项目 -->

## 使用指南

<!-- 描述如何使用该项目 -->

## 如何贡献

<!-- 告诉其他开发者如果给该项目贡献源码 -->


## 社区(可选)

<!-- 如果有需要可以介绍一些社区相关的内容 -->

## 关于作者

<!-- 这里写上项目作者 -->

## 谁在用(可选)

<!-- 可以列出使用本项目的其他有影响力的项目，算是给项目打个广告吧 -->

## 许可证

<!-- 这里链接上该项目的开源许可证 -->
```

更具体的示例，可以参考 IAM 系统的 README.md 文件 。 

这里，有个在线的 README 生成工具，可以参考下：readme.so。 

#### 项目文档规范

项目文档包括一切需要文档化的内容，它们通常集中放在 /docs 目录下。当创建团队的项目文档时，通常会预先规划并创建好一些目录，用来存放不同的文档。因此，在开始 Go 项目开发之前，也要制定一个软件文档规范。

好的文档规范有 2 个优点：易读和可以快速定位文档。 

不同项目有不同的文档需求，在制定文档规范时，可以考虑包含两类文档。

- 开发文档：用来说明该项目的开发流程，比如如何搭建开发环境、构建二进制文件、测试、部署等。 
- 用户文档：软件的使用文档，对象一般是软件的使用者，内容可根据需要添加。比如，可以包括 API 文档、SDK 文档、安装文档、功能介绍文档、最佳实践、操作指南、常见问题等。

为了方便全球开发者和用户使用，开发文档和用户文档，可以预先规划好英文和中文 2 个版本。 

为了加深理解，这里来看下实战项目的文档目录结构：

```bash
docs
├── devel # 开发文档，可以提前规划好，英文版文档和中文版文档
│ 	├── en-US/ # 英文版文档，可以根据需要组织文件结构
│ 	└── zh-CN # 中文版文档，可以根据需要组织文件结构
│ 		└── development.md # 开发手册，可以说明如何编译、构建、运行项目
├── guide # 用户文档
│ ├── en-US/ # 英文版文档，可以根据需要组织文件结构
│ └── zh-CN # 中文版文档，可以根据需要组织文件结构
│ 	├── api/ # API文档
│ 	├── best-practice # 最佳实践，存放一些比较重要的实践文章
│ 	│	 	└── authorization.md
│ 	├── faq # 常见问题
│ 	│ 	├── iam-apiserver
│ 	│ 	└── installation
│ 	├── installation # 安装文档
│ 	│ 	└── installation.md
│ 	├── introduction/ # 产品介绍文档
│	  ├── operation-guide # 操作指南，里面可以根据RESTful资源再划分为更细的子目录，用来存放系统核心/全部功能的操作手册
│ 	│  ├── policy.md
│ 	│  ├── secret.md
│ 	│  └── user.md
│ 	├── quickstart # 快速入门
│ 	│  └── quickstart.md
│ 	├── README.md # 用户文档入口文件
│ 	└── sdk # SDK文档
│ 		 └── golang.md
└── images # 图片存放目录
		└── 部署架构v1.png
```

#### API 接口文档规范

接口文档又称为 API 文档，一般由后台开发人员编写，用来描述组件提供的 API 接口，以及如何调用这些 API 接口。

在项目初期，接口文档可以解耦前后端，让前后端并行开发：前端只需要按照接口文档实现调用逻辑，后端只需要按照接口文档提供功能。 

当前后端都开发完成之后，就可以直接进行联调，提高开发效率。在项目后期，接口文档可以提供给使用者，不仅可以降低组件的使用门槛，还能够减少沟通成本。 

显然，一个有固定格式、结构清晰、内容完善的接口文档，就非常重要了。那么该如何编写接口文档，它又有什么规范呢？ 

接口文档有四种编写方式，包括编写 Word 格式文档、借助工具编写、通过注释生成和编写 Markdown 格式文档。具体的实现方式见下表：

![image-20211104223838561](IAM-document.assets/image-20211104223838561.png)

其中，通过注释生成和编写 Markdown 格式文档这 2 种方式用得最多。采用编写 Markdown 格式文档的方式，原因如下：

- 相比通过注释生成的方式，编写 Markdown 格式的接口文档，能表达更丰富的内容和格式，不需要在代码中添加大量注释。 
- 相比 Word 格式的文档，Markdown 格式文档占用的空间更小，能够跟随代码仓库一起发布，方便 API 文档的分发和查找。 
- 相比在线 API 文档编写工具，Markdown 格式的文档免去了第三方平台依赖和网络的限 制。

API 接口文档又要遵循哪些规范呢？

- 其实，一个规范的 API 接口文档，通常需要包含一个完整的 API 接口介绍文档、API 接口变更历史文档、通用说明、数据结构说明、错误码描述和 API 接口使用文档。
- API 接口使用文档中需要包含接口描述、请求方法、请求参数、 输出参数和请求示例。 

当然，根据不同的项目需求，API 接口文档会有不同的格式和内容。以实战项目采用的 API 接口文档规范为例，接口文档拆分为以下几个 Markdown 文件，并存放在目录 docs/guide/zh-CN/api 中：

- README.md ：API 接口介绍文档，会分类介绍 IAM 支持的 API 接口，并会存放相关 API 接口文档的链接，方便开发者查看。
- CHANGELOG.md ：API 接口文档变更历史，方便进行历史回溯，也可以使调用者决定是否进行功能更新和版本更新。
- generic.md ：用来说明通用的请求参数、返回参数、认证方法和请求方法等。
- struct.md ：用来列出接口文档中使用的数据结构。这些数据结构可能被多个 API 接口使用，会在 user.md、secret.md、policy.md 文件中被引用。 
- user.md 、 secret.md 、 policy.md ：API 接口文档，相同 REST 资源的接口 会存放在一个文件中，以 REST 资源名命名文档名。 
- error_code.md ：错误码描述，通过程序自动生成。

这里拿 user.md 接口文档为例，解释下接口文档是如何写的。user.md 文件记录 了用户相关的接口，每个接口按顺序排列，包含如下 5 部分。

- 接口描述：描述接口实现了什么功能。 
- 请求方法：接口的请求方法，格式为 HTTP 方法请求路径，例如 POST /v1/users。 在通用说明中的请求方法部分，会说明接口的请求协议和请求地址。
- 输入参数：接口的输入字段，它又分为 Header 参数、Query 参数、Body 参数、Path 参数。每个字段通过：参数名称、必选、类型和描述 4 个属性来描述。如果参数有限制或者默认值，可以在描述部分注明。 
- 输出参数：接口的返回字段，每个字段通过参数名称、类型 和 描述 3 个属性来描述。 
- 请求示例：一个真实的 API 接口请求和返回示例。



### 版本规范

在做 Go 项目开发时，建议把所有组件都加入版本机制。原因主要有两个：

- 一是通过版本号，可以很明确地知道组件是哪个版本，从而定位到该组件的功能和代码，方便我们定位问题。
- 二是发布组件时携带版本号，可以让使用者知道目前的项目进度，以及使用版本和上一个版本的功能差别等。 

目前业界主流的版本规范是语义化版本规范，也是 IAM 系统采用的版本规范。那什么是语义化版本规范呢？ 

#### 什么是语义化版本规范（SemVer）？ 

语义化版本规范（SemVer，Semantic Versioning）是 GitHub 起草的一个具有指导意义的、统一的版本号表示规范。它规定了版本号的表示、增加和比较方式，以及不同版本号代表的含义。 

在这套规范下，版本号及其更新方式包含了相邻版本间的底层代码和修改内容的信息。语义化版本格式为：`主版本号.次版本号.修订号（X.Y.Z）`，其中 X、Y 和 Z 为非负的整数，且禁止在数字前方补零。 

版本号可按以下规则递增：

- 主版本号（MAJOR）：当做了不兼容的 API 修改。 
- 次版本号（MINOR）：当做了向下兼容的功能性新增及修改。这里有个不成文的约定需要注意，偶数为稳定版本，奇数为开发版本。 
- 修订号（PATCH）：当做了向下兼容的问题修正。

例如，v1.2.3 是一个语义化版本号，版本号中每个数字的具体含义见下图：

![image-20211104225538739](IAM-document.assets/image-20211104225538739.png)

可能还看过这么一种版本号：v1.2.3-alpha。这其实是把先行版本号（Pre-release） 和版本编译元数据，作为延伸加到了`主版本号.次版本号.修订号`的后面，格式为 `X.Y.Z[-先行版本号][+版本编译元数据]`，如下图所示：

![image-20211104225709347](IAM-document.assets/image-20211104225709347.png)

来分别看下先行版本号和版本编译元数据是什么意思。 

先行版本号意味着，该版本不稳定，可能存在兼容性问题，格式为：`X.Y.Z-[一连串以句 点分隔的标识符] `，比如下面这几个例子：

```bash
1.0.0-alpha
1.0.0-alpha.1
1.0.0-0.3.7
1.0.0-x.7.z.92
```

编译版本号，一般是编译器在编译过程中自动生成的，只定义其格式，并不进行人为控制。下面是一些编译版本号的示例：

```bash
1.0.0-alpha+001
1.0.0+20130313144700
1.0.0-beta+exp.sha.5114f85
```

注意，先行版本号和编译版本号只能是字母、数字，且不可以有空格。 

#### 语义化版本控制规范

语义化版本控制规范比较多，这里介绍几个比较重要的。

- 标记版本号的软件发行后，禁止改变该版本软件的内容，任何修改都必须以新版本发行。 
- 主版本号为零（0.y.z）的软件处于开发初始阶段，一切都可能随时被改变，这样的公共 API 不应该被视为稳定版。1.0.0 的版本号被界定为第一个稳定版本，之后的所有版本号更新都基于该版本进行修改。 
- 修订号 Z（x.y.Z | x > 0）必须在只做了向下兼容的修正时才递增，这里的修正其实就是 Bug 修复。 
- 次版本号 Y（x.Y.z | x > 0）必须在有向下兼容的新功能出现时递增，在任何公共 API 的功能被标记为弃用时也必须递增，当有改进时也可以递增。其中可以包括修订级别的改变。每当次版本号递增时，修订号必须归零。 
- 主版本号 X（X.y.z | X > 0）必须在有任何不兼容的修改被加入公共 API 时递增。其中可以包括次版本号及修订级别的改变。每当主版本号递增时，次版本号和修订号必须归零。

#### 如何确定版本号？

说了这么多，到底该如何确定版本号呢？ 

这里总结了这么几个经验： 

- 第一，在实际开发的时候，建议使用 0.1.0 作为第一个开发版本号，并在后续的每次发行时递增次版本号。 
- 第二，当版本是一个稳定的版本，并且第一次对外发布时，版本号可以定为 1.0.0。 
- 第三，当严格按照 Angular commit message 规范提交代码时，版本号可以这么来确定：
  - fix 类型的 commit 可以将修订号 +1。 
  - feat 类型的 commit 可以将次版本号 +1。 
  - 带有 BREAKING CHANGE 的 commit 可以将主版本号 +1。

### 总结

一套好的规范，就是一个项目开发的“规矩”，它可以确保整个项目的可维护性、可阅读性，减少 Bug 数等。 

一个项目的规范设计主要包括编码类和非编码类这两类规范。学习了开源规范、文档规范和版本规范，回顾重点内容。

- 新开发的项目最好按照开源标准来规范，以驱动其成为一个高质量的项目。 
- 开发之前，最好提前规范好文档目录，并选择一种合适的方式来编写 API 文档。在实战项目中，采用的是 Markdown 格式，也推荐使用这种方式。 
- 项目要遵循版本规范，目前业界主流的版本规范是语义化版本规范，也是推荐的版本规范。



### 课后练习

1. 除了今天介绍的这些非编码类规范之外，在开发中还用到过哪些规范? 
2. 试着用介绍的 API 文档规范，书写一份当前项目的 API 接口。



## 规范设计之Commit规范

学习非编码类规范中的 Commit 规范。

在做代码开发时，经常需要提交代码，提交代码时需要填写 Commit Message（提交说明），否则就不允许提交。

而在实际开发中，每个研发人员提交 Commit Message 的格式可以说是五花八门，有用中文的、有用英文的，甚至有的直接填写“11111”。这样的 Commit Message，时间久了可能连提交者自己都看不懂所表述的修改内容，更别说给别人看了。 

所以在 Go 项目开发时，一个好的 Commit Message 至关重要：

- 可以使自己或者其他开发人员能够清晰地知道每个 commit 的变更内容，方便快速浏览变更历史，比如可以直接略过文档类型或者格式化类型的代码变更。 
- 可以基于这些 Commit Message 进行过滤查找，比如只查找某个版本新增的功能： git log --oneline --grep "^feat|^fix|^perf"。 
- 可以基于规范化的 Commit Message 生成 Change Log。 
- 可以依据某些类型的 Commit Message 触发构建或者发布流程，比如当 type 类型为 feat、fix 时才触发 CI 流程。 
- 确定语义化版本的版本号。比如 fix 类型可以映射为 PATCH 版本，feat 类型可以映射为 MINOR 版本。带有 BREAKING CHANGE 的 commit，可以映射为 MAJOR 版 本。通过这种方式来自动生成版本号。

总结来说，一个好的 Commit Message 规范可以使 Commit Message 的可读性更好， 并且可以实现自动化。那究竟如何写一个易读的 Commit Message 呢？ 

接下来，看下如何规范 Commit Message。另外，除了 Commit Message 之外， 还会介绍跟 Commit 相关的 3 个重点，以及如何通过自动化流程来保证 Commit Message 的规范化。 

### Commit Message 的规范有哪些？ 

毫无疑问，可以根据需要自己来制定 Commit Message 规范，但是更建议采用开源社区中比较成熟的规范。

- 一方面，可以避免重复造轮子，提高工作效率。
- 另一方面， 这些规范是经过大量开发者验证的，是科学、合理的。 

目前，社区有多种 Commit Message 的规范，例如 jQuery、Angular 等。将这些规范及其格式绘制成下面一张图片，供参考：

![image-20211105222805713](IAM-document.assets/image-20211105222805713.png)

在这些规范中，Angular 规范在功能上能够满足开发者 commit 需求，在格式上清晰易读，目前也是用得最多的。 

### Angular 规范

Angular 规范其实是一种语义化的提交规范（Semantic Commit Messages），所谓语义化的提交规范包含以下内容：

- Commit Message 是语义化的：Commit Message 都会被归为一个有意义的类型，用来说明本次 commit 的类型。 
- Commit Message 是规范化的：Commit Message 遵循预先定义好的规范，比如 Commit Message 格式固定、都属于某个类型，这些规范不仅可被开发者识别也可以被工具识别。

为了方便理解 Angular 规范，直接看一个遵循 Angular 规范的 commit 历史记录，见下图：

![image-20211105223025311](IAM-document.assets/image-20211105223025311.png)

再来看一个完整的符合 Angular 规范的 Commit Message，如下图所示：

![image-20211105223112537](IAM-document.assets/image-20211105223112537.png)

通过上面 2 张图，可以看到符合 Angular Commit Message 规范的 commit 都是有一定格式，有一定语义的。 

该怎么写出符合 Angular 规范的 Commit Message 呢？ 

在 Angular 规范中，Commit Message 包含三个部分，分别是 Header、Body 和 Footer，格式如下：

```js
<type>[optional scope]: <description>
// 空行
[optional body]
// 空行
[optional footer(s)]
```

其中，Header 是必需的，Body 和 Footer 可以省略。在以上规范中，必须用括号 () 括起来， [] 后必须紧跟冒号 ，冒号后必须紧跟空格，2 个空行也是必需的。

在实际开发中，为了使 Commit Message 在 GitHub 或者其他 Git 工具上更加易读，往往会限制每行 message 的长度。根据需要，可以限制为 50/72/100 个字符，这里将长度限制在 72 个字符以内（也有一些开发者会将长度限制为 100，可根据需要自行选择）。 以下是一个符合 Angular 规范的 Commit Message：

```sh
fix($compile): couple of unit tests for IE9
# Please enter the Commit Message for your changes. Lines starting
# with '#' will be ignored, and an empty message aborts the commit.
# On branch master
# Changes to be committed:
# ...

Older IEs serialize html uppercased, but IE9 does not...
Would be better to expect case insensitive, unfortunately jasmine does not allow to user regexps for throw expectations.

Closes #392
Breaks foo.bar api, foo.baz should be used instead
```

接下来，详细看看 Angular 规范中 Commit Message 的三个部分。

#### Header

Header 部分只有一行，包括三个字段：type（必选）、scope（可选）和 subject（必选）。 

##### type 字段

先来说 type，它用来说明 commit 的类型。为了方便记忆，把这些类型做了归纳， 主要可以归为 Development 和 Production 共两类。它们的含义是：

- Development：这类修改一般是项目管理类的变更，不会影响最终用户和生产环境的代 码，比如 CI 流程、构建方式等的修改。遇到这类修改，通常也意味着可以免测发布。 
- Production：这类修改会影响最终的用户和生产环境的代码。所以对于这种改动，一定要慎重，并在提交前做好充分的测试。

在这里列出了 Angular 规范中的常见 type 和它们所属的类别，在提交 Commit Message 的时候，一定要注意区分它的类别。

举个例子，在做 Code Review 时，如果遇到 Production 类型的代码，一定要认真 Review，因为这种类型，会影响到现网用户的使用和现网应用的功能。

![image-20211105224142396](IAM-document.assets/image-20211105224142396.png)

有这么多 type，该如何确定一个 commit 所属的 type 呢？这里可以通过下面这张图来确定。

![image-20211105224256689](IAM-document.assets/image-20211105224256689.png)

如果变更了应用代码，比如某个 Go 函数代码，那这次修改属于代码类。在代码类中，有 4 种具有明确变更意图的类型：feat、fix、perf 和 style；如果代码变更不属于这 4 类，那就全都归为 refactor 类，也就是优化代码。

如果我们变更了非应用代码，例如更改了文档，那它属于非代码类。在非代码类中，有 3 种具有明确变更意图的类型：test、ci、docs；如果非代码变更不属于这 3 类，那就全部归入到 chore 类。 

Angular 的 Commit Message 规范提供了大部分的 type，在实际开发中，可以使用部分 type，或者扩展添加自己的 type。但无论选择哪种方式，一定要保证一个项目中的 type 类型一致。 

##### scope 字段

接下来，说说 Header 的第二个字段 scope。 

scope 是用来说明 commit 的影响范围的，它必须是名词。显然，不同项目会有不同的 scope。在项目初期，可以设置一些粒度比较大的 scope，比如可以按组件名或者功能来设置 scope；后续，如果项目有变动或者有新功能，可以再用追加的方式添加新的 scope。

这门课采用的 scope，主要是根据组件名和功能来设置的。例如，支持 apiserver、 authzserver、user 这些 scope。 

这里想强调的是，scope 不适合设置太具体的值。太具体的话，一方面会导致项目有太多的 scope，难以维护。另一方面，开发者也难以确定 commit 属于哪个具体的 scope，导致错放 scope，反而会使 scope 失去了分类的意义。 

当然了，在指定 scope 时，也需要遵循预先规划的 scope，所以要将 scope 文档化，放在类似 devel 这类文档中。这一点可以参考下 IAM 项目的 scope 文档： IAM commit message scope 。

##### subject 字段

最后，再说说 subject。 

subject 是 commit 的简短描述，必须以动词开头、使用现在时。比如，可以用 change，却不能用 changed 或 changes，而且这个动词的第一个字母必须是小写。通过这个动词，可以明确地知道 commit 所执行的操作。此外还要注意，subject 的 结尾不能加英文句号。 

#### Body

Header 对 commit 做了高度概括，可以方便查看 Commit Message。那如何知道具体做了哪些变更呢？答案就是，可以通过 Body 部分，它是对本次 commit 的更详细描述，是可选的。 

Body 部分可以分成多行，而且格式也比较自由。不过，和 Header 里的一样，它也要以动词开头，使用现在时。此外，它还必须要包括修改的动机，以及和跟上一版本相比的改动点。 

在下面给出了一个范例，可以看看：

```bash
The body is mandatory for all commits except for those of scope "docs". When the body is required it must be at least 20 characters long.
```

#### Footer

Footer 部分不是必选的，可以根据需要来选择，主要用来说明本次 commit 导致的后果。 在实际应用中，Footer 通常用来说明不兼容的改动和关闭的 Issue 列表，格式如下：

```sh
BREAKING CHANGE: <breaking change summary>
// 空行
<breaking change description + migration instructions>
// 空行
// 空行
Fixes #<issue number>
```

接下来，详细说明下这两种情况：

- 不兼容的改动：如果当前代码跟上一个版本不兼容，需要在 Footer 部分，以 BREAKING CHANG: 开头，后面跟上不兼容改动的摘要。Footer 的其他部分需要说明变动的描述、变动的理由和迁移方法，例如：

  - ```bash
    BREAKING CHANGE: isolate scope bindings definition has changed and 
    	the inject option for the directive controller injection was removed.
    	
    To migrate the code follow the example below:
    
      Before:
      
      scope: {
      	myAttr: 'attribute',
      }
      
      After:
      
      scope: {
      	myAttr: '@',
      }
      The removed `inject` wasn't generaly useful for directives so there should be no code using it.
    ```

- 关闭的 Issue 列表：关闭的 Bug 需要在 Footer 部分新建一行，并以 Closes 开头列出，例如：Closes #123。如果关闭了多个 Issue，可以这样列出：Closes #123, #432, #886。例如:

  - ```bash
    Change pause version value to a constant for image
    
    	Closes #1137
    ```



#### Revert Commit 

除了 Header、Body 和 Footer 这 3 个部分，Commit Message 还有一种特殊情况：如果当前 commit 还原了先前的 commit，则应以 revert: 开头，后跟还原的 commit 的 Header。

而且，在 Body 中必须写成 This reverts commit  ，其中 hash 是要还原的 commit 的 SHA 标识。例如：

```bash
revert: feat(iam-apiserver): add 'Host' option

This reverts commit 079360c7cfc830ea8a6e13f4c8b8114febc9b48a.
```

为了更好地遵循 Angular 规范，建议在提交代码时养成不用 git commit -m，即不用 -m 选项的习惯，而是直接用 git commit 或者 git commit -a 进入交互界面编辑 Commit Message。这样可以更好地格式化 Commit Message。

但是除了 Commit Message 规范之外，在代码提交时，还需要关注 3 个重点内容： 提交频率、合并提交和 Commit Message 修改。

### Commit 相关的 3 个重要内容 

#### 提交频率

先来看下提交频率。 提交频率在实际项目开发中，如果是个人项目，随意 commit 可能影响不大，但如果是多人开发的项目，随意 commit 不仅会让 Commit Message 变得难以理解，还会让其他研发同事觉得你不专业。因此，要规定 commit 的提交频率。 

那到底什么时候进行 commit 最好呢？ 主要可以分成两种情况。

- 一种情况是，只要对项目进行了修改，一通过测试就立即 commit。比如修复完一个 bug、开发完一个小功能，或者开发完一个完整的功能，测 试通过后就提交。
- 另一种情况是，规定一个时间，定期提交。建议代码下班前固定提交一次，并且要确保本地未提交的代码，延期不超过 1 天。这样，如果本地代码丢失，可以尽可能减少丢失的代码量。

按照上面 2 种方式提交代码，可能会觉得代码 commit 比较多，看起来比较随意。或者说，想等开发完一个完整的功能之后，放在一个 commit 中一起提交。这时候，可以在最后合并代码或者提交 Pull Request 前，执行 git rebase -i 合并之前的所有 commit。 

那么如何合并 commit 呢？接下来，详细说说。

#### 合并提交 

合并提交，就是将多个 commit 合并为一个 commit 提交。这里，建议把新的 commit 合并到主干时，只保留 2~3 个 commit 记录。那具体怎么做呢？ 

在 Git 中，主要使用 git rebase 命令来合并。git rebase 也是日后开发需要经常使用的一个命令，所以一定要掌握好它的使用方法。

##### git rebase 命令介绍 

git rebase 的最大作用是它可以重写历史。 

通常会通过 `git rebase -i <Commit ID>` 使用 git rebase 命令，-i 参数表示交互（interactive），该命令会进入到一个交互界面中，其实就是 Vim 编辑器。在该界面中，可以对里面的 commit 做一些操作，交互界面如图所示：

![image-20211105230833602](IAM-document.assets/image-20211105230833602.png)

这个交互界面会首先列出给定之前（不包括，越下面越新）的所有 commit，每个 commit 前面有一个操作命令，默认是 pick。可以选择不同的 commit，并修改 commit 前面的命令，来对该 commit 执行不同的变更操作。 

git rebase 支持的变更操作如下：

![image-20211105231221977](IAM-document.assets/image-20211105231221977.png)

在上面的 7 个命令中，squash 和 fixup 可以用来合并 commit。例如用 squash 来合并， 只需要把要合并的 commit 前面的动词，改成 squash（或者 s）即可。可以看看下面的示例：

```sh
pick 07c5abd Introduce OpenPGP and teach basic usage
s de9b1eb Fix PostChecker::Post#urls
s 3e7ee36 Hey kids, stop all the highlighting
pick fa20af3 git interactive rebase, squash, amend
```

rebase 后，第 2 行和第 3 行的 commit 都会合并到第 1 行的 commit。这个时候，提交的信息会同时包含这三个 commit 的提交信息：

```sh
# This is a combination of 3 commits.
# The first commit's message is:
Introduce OpenPGP and teach basic usage
# This is the 2ndCommit Message:
Fix PostChecker::Post#urls
# This is the 3rdCommit Message:
Hey kids, stop all the highlighting
```

如果将第 3 行的 squash 命令改成 fixup 命令：

```sh
pick 07c5abd Introduce OpenPGP and teach basic usage
s de9b1eb Fix PostChecker::Post#urls
f 3e7ee36 Hey kids, stop all the highlighting
pick fa20af3 git interactive rebase, squash, amend
```

rebase 后，还是会生成两个 commit，第 2 行和第 3 行的 commit，都合并到第 1 行的 commit。但是，新的提交信息里面，第 3 行 commit 的提交信息会被注释掉：

```sh
# This is a combination of 3 commits.
# The first commit's message is:
Introduce OpenPGP and teach basic usage
# This is the 2ndCommit Message:
Fix PostChecker::Post#urls
# This is the 3rdCommit Message:
# Hey kids, stop all the highlighting
```

除此之外，在使用 git rebase 进行操作的时候，还需要注意以下几点：

- 删除某个 commit 行，则该 commit 会丢失掉。 
- 删除所有的 commit 行，则 rebase 会被终止掉。 
- 可以对 commits 进行排序，git 会从上到下进行合并。

完整演示一遍合并提交。

##### 合并提交操作示例

假设需要研发一个新的模块：user，用来在平台里进行用户的注册、登录、注销等操作，当模块完成开发和测试后，需要合并到主干分支，具体步骤如下。 

首先，新建一个分支。需要先基于 master 分支新建并切换到 feature 分支：

```bash
$ git branch -vv   # 查看当前分支
$ git checkout -b feature/user
Switched to a new branch 'feature/user'
```

这是所有 commit 历史：

```sh
$ git log --oneline
7157e9e docs(docs): append test line 'update3' to README.md
5a26aa2 docs(docs): append test line 'update2' to README.md
55892fa docs(docs): append test line 'update1' to README.md
89651d4 docs(doc): add README.md
```

接着，在 feature/user分支进行功能的开发和测试，并遵循规范提交 commit，功能开发并测试完成后，Git 仓库的 commit 记录如下：

```sh
$ git log --oneline
4ee51d6 docs(user): update user/README.md
176ba5d docs(user): update user/README.md
5e829f8 docs(user): add README.md for user
f40929f feat(user): add delete user function
fc70a21 feat(user): add create user function
7157e9e docs(docs): append test line 'update3' to README.md
5a26aa2 docs(docs): append test line 'update2' to README.md
55892fa docs(docs): append test line 'update1' to README.md
89651d4 docs(doc): add README.md
```

可以看到提交了 5 个 commit。接下来，需要将 feature/user分支的改动合并到 master 分支，但是 5 个 commit 太多了，想将这些 commit 合并后再提交到 master 分支。 

接着，合并所有 commit。在上一步中，知道 fc70a21是 feature/user分支 的第一个 commit ID，其父 commit ID 是 7157e9e，需要将7157e9e之前的所有分支进行合并，这时可以执行：

```sh
$ git rebase -i 7157e9e

# 自己实际测试的时候，使用 git rebase -i 3de0f42
```

执行命令后，会进入到一个交互界面，在该界面中，可以将需要合并的 4 个 commit，都执行 squash 操作，如下图所示： 

![image-20211105234820798](IAM-document.assets/image-20211105234820798.png)

修改完成后执行: wq 保存，会跳转到一个新的交互页面，在该页面，可以编辑 Commit Message，编辑后的内容如下图所示：

![image-20211105234929600](IAM-document.assets/image-20211105234929600.png)

\#开头的行是 git 的注释，可以忽略掉，在 rebase 后，这些行将会消失掉。修改完成后执行:wq 保存，就完成了合并提交操作。 

除此之外，这里有 2 个点需要注意：

- git rebase -i 这里的一定要是需要合并 commit 中最旧 commit 的 父 commit ID。 
- 希望将 feature/user 分支的 5 个 commit 合并到一个 commit，在 git rebase 时，需要保证其中最新的一个 commit 是 pick 状态，这样我们才可以将其他 4 个 commit 合并进去。

然后，用如下命令来检查 commits 是否成功合并。可以看到，成功将 5 个 commit 合并成为了一个 commit：d6b17e0。

```sh
$ git log --oneline
d6b17e0 feat(user): add user module with all function implements
7157e9e docs(docs): append test line 'update3' to README.md
5a26aa2 docs(docs): append test line 'update2' to README.md
55892fa docs(docs): append test line 'update1' to README.md
89651d4 docs(doc): add README.md
```

最后，就可以将 feature 分支 feature/user 的改动合并到主干分支，从而完成新功能的开发。

```sh
$ git checkout master
$ git merge feature/user
$ git log --oneline
d6b17e0 feat(user): add user module with all function implements
7157e9e docs(docs): append test line 'update3' to README.md
5a26aa2 docs(docs): append test line 'update2' to README.md
55892fa docs(docs): append test line 'update1' to README.md
89651d4 docs(doc): add README.md
```

这里给一个小提示，如果有太多的 commit 需要合并，那么可以试试这种方式：先撤销过去的 commit，然后再建一个新的。

```sh
$ git reset HEAD~3
$ git add .
$ git commit -am "feat(user): add user resource"
```

需要说明一点：除了 commit 实在太多的时候，一般情况下不建议用这种方法，有点粗暴，而且之前提交的 Commit Message 都要重新整理一遍。 

#### 修改 Commit Message 

即使有了 Commit Message 规范，但仍然可能会遇到提交的 Commit Message 不符合规范的情况，这个时候就需要能够修改之前某次 commit 的 Commit Message。 

具体来说，有两种修改方法，分别对应两种不同情况：

- git commit --amend：修改最近一次 commit 的 message；
- git rebase -i：修改某次 commit 的 message。

接下来，分别来说这两种方法。 

##### git commit --amend：修改最近一次 commit 的 message 

有时候，刚提交完一个 commit，但是发现 commit 的描述不符合规范或者需要纠正，这时候，可以通过 git commit --amend 命令来修改刚刚提交 commit 的 Commit Message。

具体修改步骤如下：

1. 查看当前分支的日志记录。

```sh
$ git log --oneline
418bd4 docs(docs): append test line 'update$i' to README.md
89651d4 docs(doc): add README.md
```

可以看到，最近一次的 Commit Message 是 `docs(docs): append test line 'update$i' to README.md`，其中 `update$i` 正常应该是 update1。

2. 更新最近一次提交的 Commit Message

在当前 Git 仓库下执行命令：git commit --amend，会进入一个交互界面，在交互界面中，修改最近一次的 Commit Message，如下图所示：

![image-20211106000657492](IAM-document.assets/image-20211106000657492.png)

修改完成后执行:wq 保存，退出编辑器之后，会在命令行显示，该 commit 的 message 的更新结果如下：

```sh
[master 55892fa] docs(docs): append test line 'update1' to README.md
Date: Fri Sep 18 13:40:42 2020 +0800
1 file changed, 1 insertion(+)
```

3. 查看最近一次的 Commit Message 是否被更新

```sh
$ git log --oneline
55892fa docs(docs): append test line 'update1' to README.md
89651d4 docs(doc): add README.md
```

可以看到最近一次 commit 的 message 成功被修改为期望的内容。 

##### git rebase -i：修改某次 commit 的 message 

如果想修改的 Commit Message 不是最近一次的 Commit Message，可以通过 git rebase -i <父 commit ID>命令来修改。这个命令在实际开发中使用频率比较高，一定要掌握。

具体来说，使用它主要分为 4 步。

1. 查看当前分支的日志记录。

```sh
$ git log --oneline
1d6289f docs(docs): append test line 'update3' to README.md
a38f808 docs(docs): append test line 'update$i' to README.md
55892fa docs(docs): append test line 'update1' to README.md
89651d4 docs(doc): add README.md
```

可以看到倒数第 2 次提交的 Commit Message 是：`docs(docs): append test line 'update$i' to README.md`，其中 update$i 正常应该是 update2。

2. 修改倒数第 2 次提交 commit 的 message。

在 Git 仓库下直接执行命令 git rebase -i 55892fa，然后会进入一个交互界面。在交互界面中，修改最近一次的 Commit Message。这里使用 reword 或者 r，保留倒数第二次的变更信息，但是修改其 message，如下图所示：

![image-20211106001358691](IAM-document.assets/image-20211106001358691.png)

修改完成后执行:wq 保存，还会跳转到一个新的交互页面，如下图所示：

![image-20211106001621276](IAM-document.assets/image-20211106001621276.png)

修改完成后执行:wq 保存，退出编辑器之后，会在命令行显示该 commit 的 message 的更新结果：

```sh
[detached HEAD 5a26aa2] docs(docs): append test line 'update2' to README.md
Date: Fri Sep 18 13:45:54 2020 +0800
1 file changed, 1 insertion(+)
Successfully rebased and updated refs/heads/master.
```

Successfully rebased and updated refs/heads/master.说明 rebase 成功， 其实这里完成了两个步骤：更新 message，更新该 commit 的 HEAD 指针。 

注意：这里一定要传入想要变更 Commit Message 的父 commit ID：`git rebase -i <父 commit ID>`。

3. 查看倒数第 2 次 commit 的 message 是否被更新。

```sh
$ git log --oneline
7157e9e docs(docs): append test line 'update3' to README.md
5a26aa2 docs(docs): append test line 'update2' to README.md
55892fa docs(docs): append test line 'update1' to README.md
89651d4 docs(doc): add README.md
```

可以看到，倒数第 2 次 commit 的 message 成功被修改为期望的内容。 

这里有两点需要注意：

- Commit Message 是 commit 数据结构中的一个属性，如果 Commit Message 有变 更，则 commit ID 一定会变，git commit --amend 只会变更最近一次的 commit ID，但是 git rebase -i 会变更父 commit ID 之后所有提交的 commit ID。 
- 如果当前分支有未 commit 的代码，需要先执行 git stash 将工作状态进行暂存，当修改完成后再执行 git stash pop 恢复之前的工作状态。

### Commit Message 规范自动化

其实，到这里也就意识到了一点：Commit Message 规范如果靠文档去约束，就会严 重依赖开发者的代码素养，并不能真正保证提交的 commit 是符合规范的。 

那么，有没有一种方式可以确保提交的 Commit Message 一定是符合规范的呢？有的，可以通过一些工具，来自动化地生成和检查 Commit Message 是否符合规范。 

另外，既然 Commit Message 是规范的，那么我们能不能利用这些规范来实现一些更酷的功能呢？答案是有的，将围绕着 Commit Message 实现的一些自动化功能绘制成了下面一张图。

![image-20211106002324783](IAM-document.assets/image-20211106002324783.png)

这些自动化功能可以分为以下 2 类：

- Commit Message 生成和检查功能：生成符合 Angular 规范的 Commit Message、 Commit Message 提交前检查、历史 Commit Message 检查。 
- 基于 Commit Message 自动生成 CHANGELOG 和 SemVer 的工具。

可以通过下面这 5 个工具自动的完成上面的功能：

- commitizen-go：使你进入交互模式，并根据提示生成 Commit Message，然后提交。 
- commit-msg：githooks，在 commit-msg 中，指定检查的规则，commit-msg 是个脚本，可以根据需要自己写脚本实现。这门课的 commit-msg 调用了 go-gitlint 来进 行检查。 
- go-gitlint：检查历史提交的 Commit Message 是否符合 Angular 规范，可以将该 工具添加在 CI 流程中，确保 Commit Message 都是符合规范的。 
- gsemver：语义化版本自动生成工具。 
- git-chglog：根据 Commit Message 生成 CHANGELOG。



### 总结 

介绍了 Commit Message 规范，主要讲了业界使用最多的 Angular 规范。 

Angular 规范中，Commit Message 包含三个部分：Header、Body 和 Footer。Header 对 commit 做了高度概括，Body 部分是对本次 commit 的更详细描述，Footer 部分主要用来说明本次 commit 导致的后果。格式如下：

```sh
<type>[optional scope]: <description>
// 空行
[optional body]
// 空行
[optional footer(s)]
```

另外，也需要控制 commit 的提交频率，比如可以在开发完一个功能、修复完一个 bug、下班前提交 commit。 

最后，也需要掌握一些常见的提交操作，例如通过 git rebase -i 来合并提交 commit，通过 git commit --amend 或 git rebase -i 来修改 commit message。 



### 课后练习

- 新建一个 git repository，提交 4 个符合 Angular 规范的 Commit Message，并合并前 2 次提交。 
- 使用 git-chglog 工具来生成 CHANGEOG，使用 gsemver 工具来生成语义化版本号。



## 目录结构设计























