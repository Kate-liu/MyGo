# IAM Document

##  IAM系统概述

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

在使用模块的时候，` $GOPATH` 是无意义的，不过它还是会把下载的依赖储存在 `$GOPATH/pkg/mod`目录中，也会把 go install 的二进制文件存放在 `$GOPATH/bin`目录中。所以，还要将 `$GOPATH/bin`、`$GOROOT/bin`加入到Linux可执行文件搜索路径中。这样，就可以直接在 bash shell 中执行 go 自带的命令，以及通过 go install 安装的命令。

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

先配置 nvim 的别名为 vi，这样当执行 vi 时，Linux 系统就会默认调用 nvim。同时，配置 EDITOR 环境变量，可以使一些工具，例如 Git 默认使用 nvim。配置方法如下：

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

为了方便，可以直接拷贝已经打包好的 Go 工具到指定目录下：

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
$ git branch -avv  # 查看
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



## 规范设计之目录结构设计

目录结构是一个项目的门面。很多时候，根据目录结构就能看出开发者对这门语言的掌握程度。所以，遵循一个好的目录规范，把代码目录设计得可维护、可扩展，甚至比文档规范、Commit 规范来得更加重要。 

那具体怎么组织一个好的代码目录呢？从 2 个维度来解答这个问题。 

- 首先，介绍组织目录的一些基本原则，这些原则可以指导你去组织一个好的代码目录。
- 然后，会介绍一些具体的、优秀的目录结构。可以通过学习它们，提炼总结出自己的目录结构设计方法，或者直接用它们作为自己的目录结构规范，也就是说结构即规范。

### 如何规范目录？ 

想设计好一个目录结构，首先要知道一个好的目录长什么样，也就是目录规范中包含哪些内容。 

目录规范，通常是指我们的项目由哪些目录组成，每个目录下存放什么文件、实现什么功能，以及各个目录间的依赖关系是什么等。一个好的目录结构至少要满足以下几个要求。

- 命名清晰：目录命名要清晰、简洁，不要太长，也不要太短，目录名要能清晰地表达出该目录实现的功能，并且目录名最好用单数。一方面是因为单数足以说明这个目录的功能，另一方面可以统一规范，避免单复混用的情况。 
- 功能明确：一个目录所要实现的功能应该是明确的、并且在整个项目目录中具有很高的辨识度。也就是说，当需要新增一个功能时，能够非常清楚地知道把这个功能放在哪个目录下。 
- 全面性：目录结构应该尽可能全面地包含研发过程中需要的功能，例如文档、脚本、源码管理、API 实现、工具、第三方包、测试、编译产物等。 
- 可预测性：项目规模一定是从小到大的，所以一个好的目录结构应该能够在项目变大时，仍然保持之前的目录结构。 
- 可扩展性：每个目录下存放了同类的功能，在项目变大时，这些目录应该可以存放更多同类功能。举个例子，有如下目录结构：

```bash
$ ls internal/
app pkg README.md
```

internal 目录用来实现内部代码，app 和 pkg 目录下的所有文件都属于内部代码。如果 internal 目录不管项目大小，永远只有 2 个文件 app 和 pkg，那么就说明 internal 目录是不可扩展的。 

相反，如果 internal 目录下直接存放每个组件的源码目录（一个项目可以由一个或多个组件组成），当项目变大、组件增多时，可以将新增加的组件代码存放到 internal 目录，这时 internal 目录就是可扩展的。例如：

```bash
$ ls internal/
apiserver authzserver iamctl pkg pump
```

刚才讲了目录结构的总体规范，现在来看 2 个具体的、可以作为目录规范的目录结构。 

通常，根据功能，可以将目录结构分为结构化目录结构和平铺式目录结构两种。

- 结构化目录结构主要用在 Go 应用中，相对来说比较复杂；
- 而平铺式目录结构主要用在 Go 包中，相对来说比较简单。 

因为平铺式目录结构比较简单，所以接下来先介绍它。 

### 平铺式目录结构 

一个 Go 项目可以是一个应用，也可以是一个代码框架 / 库，当项目是代码框架 / 库时， 比较适合采用平铺式目录结构。 

平铺方式就是在项目的根目录下存放项目的代码，整个目录结构看起来更像是一层的，这种方式在很多框架 / 库中存在，使用这种方式的好处是引用路径长度明显减少，比如 github.com/marmotedu/log/pkg/options，可缩短为 github.com/marmotedu/log/options。例如 log 包 github.com/golang/glog 就是平铺式的，目录如下：

```bash
$ ls glog/
glog_file.go glog.go glog_test.go LICENSE README

# glog 链接：https://github.com/golang/glog
```

接下来，来学习结构化目录结构，它比较适合 Go 应用，也比较复杂。 

### 结构化目录结构

当前 Go 社区比较推荐的结构化目录结构是 project-layout 。虽然它并不是官方和社区的规范，但因为组织方式比较合理，被很多 Go 开发人员接受。所以，可以把它当作是一个事实上的规范。

首先，来看下在开发一个 Go 项目时，通常应该包含的功能。这些功能内容比较多， 放在了 GitHub 的 Go 项目通常包含的功能 里，设计的目录结构应该能够包含这些功能。 

结合 project-layout，以及上面列出的 Go 项目常见功能，总结出了一套 Go 的代码结构组织方式，也就是 IAM 项目使用的目录结构。这种方式保留了 project-layout 优势的同时，还加入了一些个人的理解，希望提供一个拿来即用的目录结构规范。 

接下来，一起看看这门课的实战项目所采用的 Go 目录结构。因为实战项目目录比较多，这里只列出了一些重要的目录和文件，可以快速浏览以加深理解。

```bash
├── CHANGELOG
├── CONTRIBUTING.md
├── LICENSE
├── Makefile
├── OWNERS
├── README.md
├── SECURITY.md
├── api
│   ├── openapi
│   └── swagger
├── build
│   ├── ci
│   ├── docker
│   │   ├── iam-apiserver
│   │   ├── iam-authz-server
│   │   ├── iam-pump
│   │   └── iamctl
│   └── package
├── cmd
│   ├── iam-apiserver
│   │   └── apiserver.go
│   ├── iam-authz-server
│   │   └── authzserver.go
│   ├── iam-pump
│   │   └── pump.go
│   └── iamctl
│       └── iamctl.go
├── configs
├── deployments
├── docs
│   ├── README.md
│   ├── devel
│   │   └── zh-CN
│   ├── guide
│   │   ├── en-US
│   │   └── zh-CN
│   ├── images
├── examples
├── githooks
├── go.mod
├── go.sum
├── init
├── internal
│   ├── apiserver
│   │   ├── app.go
│   │   ├── auth.go
│   │   ├── config
│   │   │   ├── config.go
│   │   │   └── doc.go
│   │   ├── controller
│   │   │   └── v1
│   │   │       └── user
│   │   ├── grpc.go
│   │   ├── options
│   │   │   ├── options.go
│   │   │   └── validation.go
│   │   ├── router.go
│   │   ├── run.go
│   │   ├── server.go
│   │   ├── service
│   ├── authzserver
│   │   ├── analytics
│   │   │   ├── analytics.go
│   │   │   └── analytics_options.go
│   │   ├── app.go
│   │   ├── authorization
│   ├── iamctl
│   │   ├── cmd
│   │   │   ├── completion
│   │   │   ├── user
│   │   └── util
│   ├── pkg
│   │   ├── README.md
│   │   ├── code
│   │   ├── middleware
│   │   ├── options
│   │   ├── util
│   │   └── validation
│   │       ├── doc.go
│   │       └── validation.go
│   └── pump
│       ├── analytics
│       ├── app.go
│       ├── config
│       ├── pumps
├── pkg
│   ├── util
├── scripts
│   ├── lib
│   ├── make-rules
├── test
│   └── testdata
├── third_party
│   └── forked
└── tools
```

看到这一长串目录是不是有些晕？没关系，这里一起给这个大目录分下类，然后再具体看看每一类目录的作用。

一个 Go 项目包含 3 大部分：Go 应用 、项目管理和文档。所以，项目目录也可以分为这 3 大类。同时，Go 应用又贯穿开发阶段、测试阶段和部署阶段，相应的应用类的目录，又可以按开发流程分为更小的子类。当然了，Go 项目目录中还有一些不建议的目录。所以整体来看，目录结构可以按下图所示的方式来分类：

![image-20211108003351766](IAM-document.assets/image-20211108003351766.png)

接下来就先专心走一遍每个目录、每个文件的作用，等下次组织代码目录的时候，可以再回过头来看看，那时一定会理解得更深刻。 

#### Go 应用 ：主要存放前后端代码 

首先，来说说开发阶段所涉及到的目录。开发的代码包含前端代码和后端代码， 可以分别存放在前端目录和后端目录中。

1. /web

前端代码存放目录，主要用来存放 Web 静态资源，服务端模板和单页应用（SPAs）。

2. /cmd

一个项目有很多组件，可以把组件 main 函数所在的文件夹统一放在/cmd 目录下，例如：

```bash
$ ls cmd/
gendocs geniamdocs genman genswaggertypedocs genyaml iam-apiserver iam-a
$ ls cmd/iam-apiserver/
apiserver.go
```

每个组件的目录名应该跟期望的可执行文件名是一致的。这里要保证 `/cmd/<组件名>` 目录下不要存放太多的代码，如果认为代码可以导入并在其他项目中使用，那么它应该位 于 /pkg 目录中。如果代码不是可重用的，或者不希望其他人重用它，请将该代码放到 /internal 目录中。

3. /internal

存放私有应用和库代码。如果一些代码，不希望在其他应用和库中被导入，可以将这部分代码放在/internal 目录下。 

在引入其它项目 internal 下的包时，Go 语言会在编译时报错：

```bash
An import of a path containing the element “internal” is disallowed if the importing code is outside the tree rooted at the parent of the "internal" directory.
```

可以通过 Go 语言本身的机制来约束其他项目 import 项目内部的包。/internal 目录建议包含如下目录：

- /internal/apiserver：该目录中存放真实的应用代码。这些应用的共享代码存放在/internal/pkg 目录下。 
- /internal/pkg：存放项目内可共享，项目外不共享的包。这些包提供了比较基础、通用的功能，例如工具、错误码、用户验证等功能。

一开始将所有的共享代码存放在 /internal/pkg 目录下，当该共享代码做好了对外开发的准备后，再转存到/pkg目录下。 

下面，详细介绍下 IAM 项目的 internal目录 ，来加深你对 internal 的理解，目录 结构如下：

```bash
|── apiserver
│ ├── api
│ │ └── v1
│ │ 	└── user
│ ├── options
│ ├── config
│ ├── service
│ │ └── user.go
│ ├── store
│ │ ├── mysql
│ │ │ └── user.go
│ │ ├── fake
│ └── testing
├── authzserver
│ ├── api
│ │ └── v1
│ ├── options
│ ├── store
│ └── testing
├── iamctl
│ ├── cmd
│ │ ├── cmd.go
│ │ ├── info
└── pkg
  ├── code
  ├── middleware
  ├── options
  └── validation
```

/internal 目录大概分为 3 类子目录：

- /internal/pkg：内部共享包存放的目录。 
- /internal/authzserver、/internal/apiserver、/internal/pump、/internal/apiserver ：应用目录，里面包含应用程序的实现代码。 
- /internal/iamctl：对于一些大型项目，可能还会需要一个客户端工具。

在每个应用程序内部，也会有一些目录结构，这些目录结构主要根据功能来划分：

- /internal/apiserver/api/v1：HTTP API 接口的具体实现，主要用来做 HTTP 请求的解包、参数校验、业务逻辑处理、返回。注意这里的业务逻辑处理应该是轻量级的，如果 业务逻辑比较复杂，代码量比较多，建议放到 /internal/apiserver/service 目录下。该 源码文件主要用来串流程。
-  /internal/apiserver/options：应用的 command flag。 
- /internal/apiserver/config：根据命令行参数创建应用配置。 
- /internal/apiserver/service：存放应用复杂业务处理代码。 
- /internal/apiserver/store/mysql：一个应用可能要持久化的存储一些数据，这里主要存放跟数据库交互的代码，比如 Create、Update、Delete、Get、List 等。

/internal/pkg 目录存放项目内可共享的包，通常可以包含如下目录：

- /internal/pkg/code：项目业务 Code 码。
- /internal/pkg/validation：一些通用的验证函数。 
- /internal/pkg/middleware：HTTP 处理链。

4. /pkg

/pkg 目录是 Go 语言项目中非常常见的目录，几乎能够在所有知名的开源项目（非框架）中找到它的身影，例如 Kubernetes、Prometheus、Moby、Knative 等。 

该目录中存放可以被外部应用使用的代码库，其他项目可以直接通过 import 导入这里的代码。所以，在将代码库放入该目录时一定要慎重。

5. /vendor

项目依赖，可通过 go mod vendor 创建。需要注意的是，如果是一个 Go 库，不要提交 vendor 依赖包。

6. /third_party

外部帮助工具，分支代码或其他第三方应用（例如 Swagger UI）。比如fork 了一个 第三方 go 包，并做了一 些小的改动，可以放在目录 /third_party/forked 下。一方面可以很清楚的知道该包是 fork 第三方的，另一方面又能够方便地和 upstream 同步。 

#### Go 应用：主要存放测试相关的文件和代码 

接着，再来看下测试阶段相关的目录，它可以存放测试相关的文件。

7. /test

用于存放其他外部测试应用和测试数据。/test 目录的构建方式比较灵活：对于大的项目， 有一个数据子目录是有意义的。例如，如果需要 Go 忽略该目录中的内容，可以使用 /test/data 或 /test/testdata 目录。 

需要注意的是，Go 也会忽略以“.”或 “_” 开头的目录或文件。这样在命名测试数据目录方面，可以具有更大的灵活性。 

#### Go 应用：存放跟应用部署相关的文件 

接着，再来看下与部署阶段相关的目录，这些目录可以存放部署相关的文件。

8. /configs

这个目录用来配置文件模板或默认配置。例如，可以在这里存放 confd 或 consul-template 模板文件。这里有一点要注意，配置中不能携带敏感信息，这些敏感信息，可以用占位符来替代，例如：

```bash
apiVersion: v1
user:
username: ${CONFIG_USER_USERNAME} # iam 用户名
password: ${CONFIG_USER_PASSWORD} # iam 密码
```

9. /deployments

用来存放 Iaas、PaaS 系统和容器编排部署配置和模板（Docker-Compose， Kubernetes/Helm，Mesos，Terraform，Bosh）。在一些项目，特别是用 Kubernetes 部署的项目中，这个目录可能命名为 deploy。 

为什么要将这类跟 Kubernetes 相关的目录放到目录结构中呢？主要是因为当前软件部署基本都在朝着容器化的部署方式去演进。

10. /init

存放初始化系统（systemd，upstart，sysv）和进程管理配置文件（runit， supervisord）。比如 sysemd 的 unit 文件。这类文件，在非容器化部署的项目中会用到。 

#### 项目管理：存放用来管理 Go 项目的各类文件 

在做项目开发时，还有些目录用来存放项目管理相关的文件，这里一起来看下。

11. /Makefile

虽然 Makefile 是一个很老的项目管理工具，但它仍然是最优秀的项目管理工具。所以，一 个 Go 项目在其根目录下应该有一个 Makefile 工具，用来对项目进行管理，Makefile 通常用来执行静态代码检查、单元测试、编译等功能。其他常见功能，你可以参考这里： Makefile 常见管理内容 。 

还有一条建议：直接执行 make 时，执行如下各项 format -> lint -> test -> build，如果是有代码生成的操作，还可能需要首先 生成代码 gen -> format -> lint -> test -> build。 在实际开发中，可以将一些重复性的工作自动化，并添加到 Makefile 文件中统一管理。

12. /scripts

该目录主要用来存放脚本文件，实现构建、安装、分析等不同功能。不同项目，里面可能存放不同的文件，但通常可以考虑包含以下 3 个目录：

- /scripts/make-rules：用来存放 makefile 文件，实现 /Makefile 文件中的各个功能。 Makefile 有很多功能，为了保持它的简洁，建议将各个功能的具体实现放 在/scripts/make-rules 文件夹下。 
- /scripts/lib：shell 库，用来存放 shell 脚本。一个大型项目中有很多自动化任务，比如发布、更新文档、生成代码等，所以要写很多 shell 脚本，这些 shell 脚本会有一些通用功能，可以抽象成库，存放在/scripts/lib 目录下，比如 logging.sh，util.sh 等。 
- /scripts/install：如果项目支持自动化部署，可以将自动化部署脚本放在此目录下。如果部署脚本简单，也可以直接放在 /scripts 目录下。

另外，shell 脚本中的函数名，建议采用语义化的命名方式，例如 iam::log::info 这种语义化的命名方式，可以使调用者轻松的辨别出函数的功能类别，便于函数的管理和引用。在 Kubernetes 的脚本中，就大量采用了这种命名方式。

13. /build

这里存放安装包和持续集成相关的文件。这个目录下有 3 个大概率会使用到的目录，在设 计目录结构时可以考虑进去。

- /build/package：存放容器（Docker）、系统（deb, rpm, pkg）的包配置和脚本。 
- /build/ci：存放 CI（travis，circle，drone）的配置文件和脚本。 
- /build/docker：存放子项目各个组件的 Dockerfile 文件。

14. /tools

存放这个项目的支持工具。这些工具可导入来自 /pkg 和 /internal 目录的代码。

15. /githooks

Git 钩子。比如，可以将 commit-msg 存放在该目录。

16. /assets

项目使用的其他资源 (图片、CSS、JavaScript 等)。

17. /website

如果不使用 GitHub 页面，那么可以在这里放置项目网站相关的数据。

#### 文档：主要存放项目的各类文档 

一个项目，也包含一些文档，这些文档有很多类别，也需要一些目录来存放这些文档，一起来看下。

18. /README.md

项目的 README 文件一般包含了项目的介绍、功能、快速安装和使用指引、详细的文档链 接以及开发指引等。有时候 README 文档会比较长，为了能够快速定位到所需内容，需要添加 markdown toc 索引，可以借助工具 tocenize 来完成索引的添加。 

这里还有个建议，介绍过 README 是可以规范化的，所以这个 README 文 档，可以通过脚本或工具来自动生成。

19. /docs

存放设计文档、开发文档和用户文档等（除了 godoc 生成的文档）。推荐存放以下几个子 目录：

- /docs/devel/{en-US,zh-CN}：存放开发文档、hack 文档等。
- /docs/guide/{en-US,zh-CN}: 存放用户手册，安装、quickstart、产品文档等，分为中文文档和英文文档。 
- /docs/images：存放图片文件。

20. /CONTRIBUTING.md

如果是一个开源就绪的项目，最好还要有一个 CONTRIBUTING.md 文件，用来说明如何贡献代码，如何开源协同等等。CONTRIBUTING.md 不仅能够规范协同流程，还能降低第三方开发者贡献代码的难度。

21. /api

/api 目录中存放的是当前项目对外提供的各种不同类型的 API 接口定义文件，其中可能包 含类似 /api/protobuf-spec、/api/thrift-spec、/api/http-spec、 openapi、swagger 的目录，这些目录包含了当前项目对外提供和依赖的所有 API 文 件。例如，如下是 IAM 项目的 /api 目录：

```bash
├── openapi/
│ 	└── README.md
└── swagger/
    ├── docs/
    ├── README.md
    └── swagger.yaml
```

二级目录的主要作用，就是在一个项目同时提供了多种不同的访问方式时，可以分类存放。用这种方式可以避免潜在的冲突，也能让项目结构更加清晰。

22. /LICENSE

版权文件可以是私有的，也可以是开源的。常用的开源协议有：Apache 2.0、MIT、 BSD、GPL、Mozilla、LGPL。有时候，公有云产品为了打造品牌影响力，会对外发布一个本产品的开源版本，所以在项目规划初期最好就能规划下未来产品的走向，选择合适的 LICENSE。 

为了声明版权，可能会需要将 LICENSE 头添加到源码文件或者其他文件中，这部分工作可以通过工具实现自动化，推荐工具： addlicense 。

当代码中引用了其它开源代码时，需要在 LICENSE 中说明对其它源码的引用，这就需要知道代码引用了哪些源码，以及这些源码的开源协议，可以借助工具来进行检查，推荐工 具： glice 。至于如何说明对其它源码的引用，可以参考下 IAM 项目的 LICENSE 文件。

23. /CHANGELOG

当项目有更新时，为了方便了解当前版本的更新内容或者历史更新内容，需要将更新记录 存放到 CHANGELOG 目录。编写 CHANGELOG 是一个复杂、繁琐的工作，我们可以结合 Angular 规范 和 git-chglog 来自动生成 CHANGELOG。

24. /examples

存放应用程序或者公共包的示例代码。这些示例代码可以降低使用者的上手门槛。 

#### 不建议的目录 

除了上面这些建议的目录，在 Go 项目中，还有一些目录是不建议包含的，这些目录不符合 Go 的设计哲学。

1. /src/

一些开发语言，例如 Java 项目中会有 src 目录。在 Java 项目中， src 目录是一种常见的模式，但在 Go 项目中，不建议使用 src 目录。 

其中一个重要的原因是：在默认情况下，Go 语言的项目都会被放置到$GOPATH/src 目录 下。这个目录中存放着所有代码，如果在自己的项目中使用/src 目录，这个包的导入路径中就会出现两个 src，例如：

```bash
$GOPATH/src/github.com/marmotedu/project/src/main.go
```

这样的目录结构看起来非常怪。

2. /model

在 Go 项目里，不建议将类型定义统一存放在 model 目录中，这样做一方面不符合 Go 按功能拆分的设计哲学。

另一方面，别人在阅读代码时，可能不知道这些类型在哪里使用， 修改了结构体，也不知道有多大影响。

建议将类型定义放在它被使用的模块中。

3. xxs/

在 Go 项目中，要避免使用带复数的目录或者包。建议统一使用单数。

#### 一些建议

上面介绍的目录结构包含很多目录，但一个小型项目用不到这么多目录。对于小型项目， 可以考虑先包含 cmd、pkg、internal 3 个目录，其他目录后面按需创建，例如：

```sh
$ tree --noreport -L 2 tms
tms
├── cmd
├── internal
├── pkg
└── README.md
```

另外，在设计目录结构时，一些空目录无法提交到 Git 仓库中，但又想将这个空目录 上传到 Git 仓库中，以保留目录结构。这时候，可以在空目录下加一个 .keep 文件，例 如：

```sh
$ ls -A build/ci/
.keep
```



### 总结 

主要学习了怎么设计代码的目录结构。

- 先讲了目录结构的设计思路：在设计目录结构时，要确保目录名是清晰的，功能是明确的，并且设计的目录结构是可扩展的。
- 然后，一起学习了 2 种具体的目录结构：结构化目录结构和平铺式目录结构。
  - 结构化目录结构比较适合 Go 应用，
  - 平铺式目录结构比较适合框架 / 库。
  - 因为这 2 种目录结构组织比较合理，可以把它们作为目录规范来使用。 



### 课后练习

1. 试着用本节描述的目录规范，重构下当前的项目，并看下有啥优缺点。 
2. 思考下工作中遇到过哪些比较好的目录结构，它们有什么优点和可以改进的地方。



## 规范设计之工作流设计

如何设计合理的开发模式。 

一个企业级项目是由多人合作完成的，不同开发者在本地开发完代码之后，可能提交到同一个代码仓库，同一个开发者也可能同时开发几个功能特性。这种多人合作开发、多功能并行开发的特性如果处理不好，就会带来诸如丢失代码、合错代码、代码冲突等问题。

所以，在编码之前，需要设计一个合理的开发模式。又因为目前开发者基本都是基于 Git 进行开发的，所以本节课，会教怎么基于 Git 设计出一个合理的开发模式。 

那么如何设计工作流呢？可以根据需要，自己设计工作流，也可以采用业界沉淀下来的、设计好的、受欢迎的工作流。

- 一方面，这些工作流经过长时间的实践，被证明是合理 的；
- 另一方面，采用一种被大家熟知且业界通用的工作流，会减少团队内部磨合的时间。

会介绍 4 种受欢迎的工作流，可以选择其中一种作为工作流设计。 

在使用 Git 开发时，有 4 种常用的工作流，也叫开发模式，按演进顺序分为集中式工作流、功能分支工作流、Git Flow 工作流和 Forking 工作流。接下来，会按演进顺序分别介绍这 4 种工作流。 

### 集中式工作流 

先来看看集中式工作流，它是最简单的一种开发方式。集中式工作流的工作模式如下图所示：

![image-20211108202057517](IAM-document.assets/image-20211108202057517.png)

A、B、C 为 3 位开发者，每位开发者都在本地有一份远程仓库的拷贝：本地仓库。

A、B、 C 在本地的 master 分支开发完代码之后，将修改后的代码 commit 到远程仓库，如果有冲突就先解决本地的冲突再提交。在进行了一段时间的开发之后，远程仓库 master 分支的日志可能如下图所示：

![image-20211108202204587](IAM-document.assets/image-20211108202204587.png)

集中式工作流是最简单的开发模式，但它的缺点也很明显：不同开发人员的提交日志混杂在一起，难以定位问题。如果同时开发多个功能，不同功能同时往 master 分支合并，代码之间也会相互影响，从而产生代码冲突。 

和其他工作流相比，集中式工作流程的代码管理较混乱，容易出问题，因此适合用在团队人数少、开发不频繁、不需要同时维护多个版本的小项目中。当想要并行开发多个功能时，这种工作流就不适用了，这时候怎么办呢？接下来看功能分支工作流。

### 功能分支工作流 

功能分支工作流基于集中式工作流演进而来。在开发新功能时，基于 master 分支新建一个功能分支，在功能分支上进行开发，而不是直接在本地的 master 分支开发，开发完成之后合并到 master 分支，如下图所示：

![image-20211108202421595](IAM-document.assets/image-20211108202421595.png)

相较于集中式工作流，这种工作流让不同功能在不同的分支进行开发，只在最后一步合并到 master 分支，不仅可以避免不同功能之间的相互影响，还可以使提交历史看起来更加简洁。 

还有，在合并到 master 分支时，需要提交 PR（pull request），而不是直接将代码 merge 到 master 分支。PR 流程不仅可以把分支代码提供给团队其他开发人员进行 CR（Code Review），还可以在 PR 页面讨论代码。

通过 CR ，可以确保合并到 master 的代码是健壮的；通过 PR 页面的讨论，可以使开发者充分参与到代码的讨论中， 有助于提高代码的质量，并且提供了一个代码变更的历史回顾途径。 那么，功能分支工作流具体的开发流程是什么呢？一起来看下。

1. 基于 master 分支新建一个功能分支，功能分支可以取一些有意义的名字，便于理解， 例如 feature/rate-limiting。

```sh
$ git checkout -b feature/rate-limiting
```

2. 在功能分支上进行代码开发，开发完成后 commit 到功能分支。

```sh
$ git add limit.go
$ git commit -m "add rate limiting"
```

3. 将本地功能分支代码 push 到远程仓库。

```sh
$ git push origin feature/rate-limiting
```

4. 在远程仓库上创建 PR（例如：GitHub）。

进入 GitHub 平台上的项目主页，点击 Compare & pull request 提交 PR，如下图所示。

![image-20211108203937993](IAM-document.assets/image-20211108203937993.png)

点击 Compare & pull request 后会进入 PR 页面，在该页面中可以根据需要填写评论， 最后点击 Create pull request 提交 PR。

5. 代码管理员收到 PR 后，可以 CR 代码，CR 通过后，再点击 Merge pull request 将 PR 合并到 master，如下图所示。

![image-20211108205844339](IAM-document.assets/image-20211108205844339.png)

图中的“Merge pull request” 提供了 3 种 merge 方法：

- Create a merge commit：GitHub 的底层操作是 git merge --no-ff。feature 分支上所有的 commit 都会加到 master 分支上，并且会生成一个 merge commit。这种方式可以让我们清晰地知道是谁做了提交，做了哪些提交，回溯历史的时候也会更加方便。 
- Squash and merge：GitHub 的底层操作是 git merge --squash。Squash and merge 会使该 pull request 上的所有 commit 都合并成一个 commit ，然后加到 master 分支上，但原来的 commit 历史会丢失。如果开发人员在 feature 分支上提交 的 commit 非常随意，没有规范，那么我们可以选择这种方法来丢弃无意义的 commit。但是在大型项目中，每个开发人员都应该是遵循 commit 规范的，因此不建议在团队开发中使用 Squash and merge。 
- Rebase and merge：GitHub 的底层操作是 git rebase。这种方式会将 pull request 上的所有提交历史按照原有顺序依次添加到 master 分支的头部（HEAD）。因为 git rebase 有风险，在不完全熟悉 Git 工作流时，不建议 merge 时选择这个。

通过分析每个方法的优缺点，在实际的项目开发中，比较推荐你使用 Create a merge commit 方式。 

从刚才讲完的具体开发流程中，可以感受到，功能分支工作流上手比较简单，不仅能使你并行开发多个功能，还可以添加 code review，从而保障代码质量。当然它也有缺点，就是无法给分支分配明确的目的，不利于团队配合。它适合用在开发团队相对固定、 规模较小的项目中。

接下来要讲的 Git Flow 工作流以功能分支工作流为基础，较好地解决了上述问题。 

### Git Flow 工作流

Git Flow 工作流是一个非常成熟的方案，也是非开源项目中最常用到的工作流。它定义了一个围绕项目发布的严格分支模型，通过为代码开发、发布和维护分配独立的分支来让项目的迭代流程更加顺畅，比较适合大型的项目或者迭代速度快的项目。接下来，会通过介绍 Git Flow 的 5 种分支和工作流程，来讲解 GIt Flow 是如何工作的。 

#### Git Flow 的 5 种分支

Git Flow 中定义了 5 种分支，分别是 master、develop、feature、release 和 hotfix。 其中，master 和 develop 为常驻分支，其他为非常驻分支，不同的研发阶段会用到不同 的分支。这 5 种分支的详细介绍见下表：

![image-20211108210333538](IAM-document.assets/image-20211108210333538.png)

#### Git Flow 开发流程 

这里用一个实际的例子来演示下 Git Flow 的开发流程。场景如下： 

- a. 当前版本为：0.9.0。 
- b. 需要新开发一个功能，使程序执行时向标准输出输出“hello world”字符串。 
- c. 在开发阶段，线上代码有 Bug 需要紧急修复。 

假设 Git 项目名为 gitflow-demo，项目目录下有 2 个文件，分别是 README.md 和 main.go，内容如下。

```go
package main

import "fmt"

func main() {
	fmt.Println("callmainfunction")
}
```

具体的开发流程有 12 步，可以跟着以下步骤操作练习。

1. 创建一个常驻的分支：develop。

```sh
$ git checkout -b develop master
```

2. 基于 develop 分支，新建一个功能分支：feature/print-hello-world。

```sh
$ git checkout -b feature/print-hello-world develop
```

3. feature/print-hello-world 分支中，在 main.go 文件中添加一行代码 fmt.Println("Hello")，添加后的代码如下。

```go
package main

import "fmt"

func main() {
	fmt.Println("callmainfunction")
	fmt.Println("Hello")
}
```

4. 紧急修复 Bug。

正处在新功能的开发中（只完成了 fmt.Println("Hello") 而非 fmt.Println("Hello World")）突然线上代码发现了一个 Bug，要立即停止手上的工作，修复线上的 Bug，步骤如下。

```sh
$ git stash # 1. 开发工作只完成了一半，还不想提交，可以临时保存修改至堆栈区
$ git checkout -b hotfix/print-error master # 2. 从 master 建立 hotfix 分支
$ vi main.go # 3. 修复 bug，callmainfunction -> call main function
$ git commit -a -m 'fix print message error bug' # 4. 提交修复
$ git checkout develop # 5. 切换到 develop 分支
$ git merge --no-ff hotfix/print-error # 6. 把 hotfix 分支合并到 develop 分支
$ git checkout master # 7. 切换到 master 分支
$ git merge --no-ff hotfix/print-error # 8. 把 hotfix 分支合并到 master
$ git tag -a v0.9.1 -m "fix log bug" # 9. master 分支打 tag
$ go build -v . # 10. 编译代码，并将编译好的二进制更新到生产环境
$ git branch -d hotfix/print-error # 11. 修复好后，删除 hotfix/xxx 分支, git branch -D hotfix/print-error
$ git checkout feature/print-hello-world # 12. 切换到开发分支下
$ git merge --no-ff develop # 13. 因为 develop 有更新，这里最好同步更新下
$ git stash pop # 14. 恢复到修复前的工作状态
```

5. 继续开发。

在 main.go 中加入 fmt.Println("Hello World")。

6. 提交代码到 feature/print-hello-world 分支。

```sh
$ git commit -a -m "print 'hello world'"
```

7. 在 feature/print-hello-world 分支上做 code review。

首先，需要将 feature/print-hello-world push 到代码托管平台，例如 GitHub 上。

```sh
$ git push origin feature/print-hello-world
```

![image-20211108212750730](IAM-document.assets/image-20211108212750730.png)

创建完 pull request 之后，就可以指定 Reviewers 进行 code review，如下图所示。

![image-20211108212846637](IAM-document.assets/image-20211108212846637.png)

8. code review 通过后，由代码仓库 matainer 将功能分支合并到 develop 分支。

```sh
$ git checkout develop
$ git merge --no-ff feature/print-hello-world
```

9. 基于 develop 分支，创建 release 分支，测试代码。

```sh
$ git checkout -b release/1.0.0 develop
$ go build -v . # 构建后，部署二进制文件，并测试
```

10. 测试失败，因为要求打印“hello world”，但打印的是“Hello World”，修复的 时候，直接在 release/1.0.0 分支修改代码，修改完成后，提交并编译部署。

```sh
$ git commit -a -m "fix bug"
$ go build -v .
```

11. 测试通过后，将功能分支合并到 master 分支和 develop 分支。

```sh
$ git checkout develop
$ git merge --no-ff release/1.0.0
$ git checkout master  # git checkout main
$ git merge --no-ff release/1.0.0
$ git tag -a v1.0.0 -m "add print hello world" # master 分支打 tag
```

12. 删除 feature/print-hello-world 分支，也可以选择性删除 release/1.0.0 分支。

```sh
$ git branch -d feature/print-hello-world

# 选择性
$ git branch -d release/1.0.0
```

亲自操作一遍之后，应该会更了解这种模式的优缺点。

- 它的缺点，就是刚才已经体会到的，它有一定的上手难度。
- 不过 Git Flow 工作流还是有很多优点的：Git Flow 工作流的每个分支分工明确，这可以最大程度减少它们之间的相互影响。因为可以创建多个分支， 所以也可以并行开发多个功能。另外，和功能分支工作流一样，它也可以添加 code review，保障代码质量。

因此，Git Flow 工作流比较适合开发团队相对固定，规模较大的项目。 

### Forking 工作流 

上面讲的 Git Flow 是非开源项目中最常用的，而在开源项目中，最常用到的是 Forking 工作流，例如 Kubernetes、Docker 等项目用的就是这种工作流。这里，先来了解下fork 操作。 

fork 操作是在个人远程仓库新建一份目标远程仓库的副本，比如在 GitHub 上操作时，在项目的主页点击 fork 按钮（页面右上角），即可拷贝该目标远程仓库。

Forking 工作流的流程如下图所示。

![image-20211108220013570](IAM-document.assets/image-20211108220013570.png)

假设开发者 A 拥有一个远程仓库，如果开发者 B 也想参与 A 项目的开发，B 可以 fork 一 份 A 的远程仓库到自己的 GitHub 账号下。后续 B 可以在自己的项目进行开发，开发完成后，B 可以给 A 提交一个 PR。这时候 A 会收到通知，得知有新的 PR 被提交，A 会去查看 PR 并 code review。如果有问题，A 会直接在 PR 页面提交评论，B 看到评论后会做进一步的修改。最后 A 通过 B 的 PR 请求，将代码合并进了 A 的仓库。这样就完成了 A 代码仓库新特性的开发。如果有其他开发者想给 A 贡献代码，也会执行相同的操作。 

GitHub 中的 Forking 工作流详细步骤共有 6 步（假设目标仓库为 gitflow-demo），可以跟着以下步骤操作练习。

1. Fork 远程仓库到自己的账号下。

访问 https://github.com/marmotedu/gitflow-demo ，点击 fork 按钮。fork 后的仓库地址为：https://github.com/colin404fork/gitflow-demo 。

2) 克隆 fork 的仓库到本地。

```sh
$ git clone https://github.com/colin404fork/gitflow-demo
$ cd gitflow-demo
$ git remote add upstream https://github.com/marmotedu/gitflow-demo
$ git remote set-url --push upstream no_push # Never push to upstream master
$ git remote -v # Confirm that your remotes make sense
origin https://github.com/colin404fork/gitflow-demo (fetch)
origin https://github.com/colin404fork/gitflow-demo (push)
upstream https://github.com/marmotedu/gitflow-demo (fetch)
upstream https://github.com/marmotedu/gitflow-demo (push)
```

3. 创建功能分支。

首先，要同步本地仓库的 master 分支为最新的状态（跟 upstream master 分支一致）。

```sh
$ git fetch upstream
$ git checkout master  # git checkout main
$ git rebase upstream/master
```

然后，创建功能分支。

```sh
$ git checkout -b feature/add-function
```

4. 提交 commit。

在 feature/add-function 分支上开发代码，开发完代码后，提交 commit。

```sh
$ git fetch upstream # commit 前需要再次同步 feature 跟 upstream/master
$ git rebase upstream/master
$ git add <file>
$ git status
$ git commit
```

分支开发完成后，可能会有一堆 commit，但是合并到主干时，往往希望只有一个 （或最多两三个）commit，这可以使功能修改都放在一个或几个 commit 中，便于后面的阅读和维护。

这个时候，可以用 git rebase 来合并和修改我们的 commit，操作如下：

```sh
$ git rebase -i origin/master
```

还有另外一种合并 commit 的简便方法，就是先撤销过去 5 个 commit，然后再建一个新的：

```sh
$ git reset HEAD~5
$ git add .
$ git commit -am "Here's the bug fix that closes #28"
$ git push --force
```

squash 和 fixup 命令，还可以当作命令行参数使用，自动合并 commit。

```sh
$ git commit --fixup
$ git rebase -i --autosquash
```

5. push 功能分支到个人远程仓库。

在完成了开发，并 commit 后，需要将功能分支 push 到个人远程代码仓库，代码如下：

```sh
$ git push -f origin feature/add-function
```

6. 在个人远程仓库页面创建 pull request。

提交到远程仓库以后，就可以创建 pull request，然后请求 reviewers 进行代码 review，确认后合并到 master。这里要注意，创建 pull request 时，base 通常选择目标远程仓库的 master 分支。 

已经讲完了 Forking 工作流的具体步骤，它有什么优缺点呢？ 

- 结合操作特点，来看看它的优点：Forking 工作流中，项目远程仓库和开发者远程仓库完全独立，开发者通过提交 Pull Request 的方式给远程仓库贡献代码，项目维护者选择性地接受任何开发者的提交，通过这种方式，可以避免授予开发者项目远程仓库的权限， 从而提高项目远程仓库的安全性，这也使得任意开发者都可以参与项目的开发。 
- 但 Forking 工作流也有局限性，就是对于职能分工明确且不对外开源的项目优势不大。 

Forking 工作流比较适用于以下三种场景：

- （1）开源项目中；
- （2）开发者有衍生出自己的衍生版的需求；
- （3）开发者不固定，可能是任意一个能访问到项目的开发者。



### 总结 

这一讲中，基于 Git 介绍了 4 种开发模式，回顾一下吧。

- 集中式工作流：开发者直接在本地 master 分支开发代码，开发完成后 push 到远端仓库 master 分支。 
- 功能分支工作流：开发者基于 master 分支创建一个新分支，在新分支进行开发，开发完成后合并到远端仓库 master 分支。 
- Git Flow 工作流：Git Flow 工作流为不同的分支分配一个明确的角色，并定义分支之间什么时候、如何进行交互，比较适合大型项目的开发。 
- Forking 工作流：开发者先 fork 项目到个人仓库，在个人仓库完成开发后，提交 pull request 到目标远程仓库，远程仓库 review 后，合并 pull request 到 master 分支。

集中式工作流是最早的 Git 工作流，功能分支工作流以集中式工作流为基础，Git Flow 工作流又是以功能分支工作流为基础，Forking 工作流在 Git Flow 工作流基础上，解耦了个人远端仓库和项目远端仓库。

每种开发模式各有优缺点，适用于不同的场景，总结在下表中：

![image-20211108223138239](IAM-document.assets/image-20211108223138239.png)

总的来说，在选择工作流时，推荐如下：

- 非开源项目采用 Git Flow 工作流。 
- 开源项目采用 Forking 工作流。

因为这门课的实战项目对于项目开发者来说是一个偏大型的非开源项目，所以采用了 Git Flow 工作流。 



### 课后练习

1. 请新建立一个项目，并参考 Git Flow 开发流程，自己操作一遍，观察每一步的操作结果。
2. 请思考下，在 Git Flow 工作流中，如果要临时解决一个 Bug，该如何操作代码仓库。



## 研发流程设计之业界标准

如何设计研发流程。 

在 Go 项目开发中，不仅要完成产品功能的开发，还要确保整个过程是高效的，代码是高质量的。这就离不开一套设计合理的研发流程了。 

而一个不合理的研发流程会带来很多问题，例如：

- 代码管理混乱。合并代码时出现合错、合丢、代码冲突等问题。 
- 研发效率低。编译、测试、静态代码检查等全靠手动操作，效率低下。甚至，因为没有标准的流程，一些开发者会漏掉测试、静态代码检查等环节。 
- 发布效率低。发布周期长，以及发布不规范造成的现网问题频发。

所以，Go 项目开发一定要设计一个合理的研发流程，来提高开发效率、减少软件维护成本。研发流程会因为项目、团队和开发模式等的不同而有所不同，但不同的研发流程依然会有一些相似点。 

那么如何设计研发流程呢？这也是看到题目中“设计”两个字后，会直接想要问的。看到这俩字，第一时间可能会觉得是通过一系列的方法论，来告诉怎么进行流程设计。

但实际情况是，项目研发流程会因为团队、项目、需求等的不同而不同，很难概括出一个方法论去设计研发流程。 

所以在这一讲中，会介绍一种业界已经设计好的、相对标准的研发流程，来展示怎么设计研发流程。通过学习它，不仅能够了解到项目研发的通用流程，而且还可以基于这个流程来优化、定制，满足自己的流程需求。

### 在设计研发流程时，需要关注哪些点？ 

在看具体的研发流程之前，需要先思考一个问题：一个好的流程应该是什么样子的？ 

虽然刚才说了，不同团队、项目、需求的研发流程不会一成不变，但为了最大限度地提高研发效能，这些不同的流程都会遵循下面这几个原则。

- 发布效率高：研发流程应该能提高发布效率，减少发布时间和人工介入的工作量。 
- 发布质量高：研发流程应该能够提高发布质量，确保发布出去的代码是经过充分测试的，并且完全避免人为因素造成的故障。 
- 迭代速度快：整个研发流程要能支持快速迭代，产品迭代速度越快，意味着产品的竞争力越强，在互联网时代越能把握先机。 
- 明确性：整个研发流程中角色的职责、使用的工具、方法和流程都应该是明确的，这可以增强流程的可执行性。 
- 流程合理：研发流程最终是供产品、开发、测试、运维等人员使用的，所以整个流程设计不能是反人类的，要能够被各类参与人员接受并执行。 
- 柔性扩展：研发流程应该是柔性且可扩展的，能够灵活变通，并适应各类场景。 
- 输入输出：研发流程中的每个阶段都应该有明确的输入和输出，这些输入和输出标志着上一个阶段的完成，下一个阶段的开始。

明确了这些关注点，就有了设计、优化研发流程的抓手了。接下来，就可以一起去学习一套业界相对标准的研发流程了。在学习的过程中，也能更好地理解对各个流程的一些经验和建议了。 

### 业界相对标准的研发流程，长啥样？ 

一个项目从立项到结项，中间会经历很多阶段。业界相对标准的划分，是把研发流程分为六个阶段，分别是需求阶段、设计阶段、开发阶段、测试阶段、发布阶段、运营阶段。

其中，开发人员需要参与的阶段有 4 个：设计阶段、开发阶段、测试阶段和发布阶段。下图就是业界相对比较标准的流程：

![image-20211109001620713](IAM-document.assets/image-20211109001620713.png)

每个阶段结束时，都需要有一个最终的产出物，可以是文档、代码或者部署组件等。这个产出物既是当前阶段的结束里程碑，又是下一阶段的输入。

所以说，各个阶段不是割裂 的，而是密切联系的整体。每个阶段又细分为很多步骤，这些步骤是需要不同的参与者去 完成的工作任务。在完成任务的过程中，可能需要经过多轮的讨论、修改，最终形成定稿。 

这里有个点一定要注意：研发流程也是一种规范，很难靠开发者的自觉性去遵守。为了让项目参与人员尽可能地遵守规范，需要借助一些工具、系统来对他们进行强约束。

所以，在设计完整的研发流程之后，需要认真思考下，有哪些地方可以实现自动化，有哪些地方可以靠工具、系统来保障规范的执行。这些自动化工具会在后续中介绍。 接下来，就具体看看研发的各个阶段，以及每个阶段的具体内容。

#### 需求阶段

需求阶段是将一个抽象的产品思路具化成一个可实施产品的阶段。

在这个阶段，产品人员会讨论产品思路、调研市场需求，并对需求进行分析，整理出一个比较完善的需求文档。 最后，产品人员会组织相关人员对需求进行评审，如果评审通过，就会进入设计阶段。 

需求阶段，一般不需要研发人员参与。但这里，还是建议研发人员积极参与产品需求的讨论。 虽然是研发，但视野和对团队的贡献，不仅仅局限在研发领域。 

这里有个点需要提醒，如果团队有测试人员，这个阶段也需要拉测试人员旁听下。 因为了解产品设计，对测试阶段测试用例的编写和功能测试等都很有帮助。 

需求阶段的产出物是一个通过评审的详细的需求文档。 

#### 设计阶段 

设计阶段，是整个产品研发过程中非常重要的阶段，包括的内容也比较多，可以看一下这张表：

![image-20211109002205503](IAM-document.assets/image-20211109002205503.png)

这里的每一个设计项都应该经过反复的讨论、打磨，最终在团队内达成共识。

这样可以确保设计是合理的，并减少返工的概率。这里想提醒的是，技术方案和实现都要经过认真讨论，并获得一致通过，否则后面因为技术方案设计不当，需要返工，要承担大部分责任。 

对于后端开发人员，在设计技术方案之前，要做好充足的调研。一个技术方案，不仅要调研业界优秀的实现，还要了解友商相同技术的实现。只有这样，才可以确保技术用最佳的方式实现。

除此之外，在这个阶段一些设计项可以并行，以缩短设计阶段的耗时。例如，产品设计和 技术设计可以并行展开。

另外，如果团队有测试人员，研发阶段最好也拉上测试人员旁听下，有利于后面的测试。 

该阶段的产出物是一系列的设计文档，这些文档会指导后面的整个研发流程。

#### 开发阶段 

开发阶段，从它的名字就知道了，这是开发人员的主战场，同时它可能也是持续时间最长的阶段。在这一阶段，开发人员根据技术设计文档，编码实现产品需求。 

开发阶段是整个项目的核心阶段，包含很多工作内容，而且每一个 Go 项目具体的步骤是不同的。把开发阶段的常见步骤总结在了下图中，有助于对它进行整体把握。

![image-20211109002518228](IAM-document.assets/image-20211109002518228.png)

来详细看下这张图里呈现的步骤。开发阶段又可以分为“开发”和“构建”两部分，先来看开发。 

首先，需要制定一个所有研发人员共同遵循的 Git 工作流规范。最常使用的是 Git Flow 工作流或者 Forking 工作流。 

为了提高开发效率，越来越多的开发者采用生成代码的方式来生成一部分代码，所以在真正编译之前可能还需要先生成代码，比如生成.pb.go 文件、API 文档、测试用例、错误码 等。建议在项目开发中，要思考怎么尽可能自动生成代码。这样不仅能提高研发效率，还能减少错误。 

对于一个开源项目，可能还需要检查新增的文件是否有版权信息。此外，根据项目不同，开发阶段还可能有其它不同的步骤。在流程的最后，通常会进行静态代码检查、单元测试和编译。编译之后，就可以启动服务，并进行自测了。

自测之后，可以遵循 Git Flow 工作流，将开发分支 push 到代码托管平台进行 code review。code review 通过之后，就可以将代码 merge 到 develop 分支上。 

接下来进入构建阶段。这一阶段最好借助 CI/CD 平台实现自动化，提高构建效率。 

合并到 develop 分支的代码同样需要进行代码扫描、单元测试，并编译打包。最后，需要进行归档，也就是将编译后的二进制文件或 Docker 镜像上传到制品库或镜像仓库。 

刚刚完整走了一遍开发阶段的常见步骤。可以看到，整个开发阶段步骤很多，而且都是高频的操作。那怎么提高效率呢？这里推荐两种方法：

- 将开发阶段的步骤通过 Makefile 实现集中管理； 
- 将构建阶段的步骤通过 CI/CD 平台实现自动化。

还需要特别注意这一点：在最终合并代码到 master（main） 之前，要确保代码是经过充分测试的。这就要求一定要借助代码管理平台提供的 Webhook 能力，在代码提交时触发 CI/CD 作业，对代码进行扫描、测试，最终编译打包，并以整个作业的成功执行作为合并代码的先决条件。 

开发阶段的产出物是满足需求的源代码、开发文档，以及编译后的归档文件。

#### 测试阶段 

测试阶段由测试工程师（也叫质量工程师）负责，这个阶段的主要流程是：测试工程师根据需求文档创建测试计划、编写测试用例，并拉研发同学一起评审测试计划和用例。评审通过后，测试工程师就会根据测试计划和测试用例对服务进行测试。 

为了提高整个研发效率，测试计划的创建和测试用例的编写可以跟开发阶段并行。 

研发人员在交付给测试时，要提供自测报告、自测用例和安装部署文档。这里要强调的是：在测试阶段，为了不阻塞测试，确保项目按时发布，研发人员应该优先解决测试同学 的 Bug，至少是阻塞类的 Bug。为了减少不必要的沟通和排障，安装部署文档要尽可能详尽和准确。

另外，也可以及时跟进测试，了解测试同学当前遇到的卡点。因为实际工作中，一些测试同学在遇到卡点时，不善于或者不会及时地同步卡点，往往研发 1 分钟就可以解决的问题，可能要花测试同学几个小时或者更久的时间去解决。 

当然，测试用例几乎不可能涵盖整个变更分支，所以对于一些难测，隐藏的测试，需要研发人员自己加强测试。 

最后，一个大特性测试完，请测试同学吃个饭吧，大家唠唠家常，联络联络感情，下次合作会更顺畅。 

测试阶段的产出物是满足产品需求、达到发布条件的源代码，以及编译后的归档文件。 

#### 发布阶段 

发布阶段主要是将软件部署上线，为了保证发布的效率和质量，需要遵循一定的发布流程，如下图所示：

![image-20211109003644303](IAM-document.assets/image-20211109003644303.png)

发布阶段按照时间线排序又分为代码发布、发布审批和服务发布 3 个子阶段。接下来，详细介绍下这 3 个子阶段。先来看一下代码发布。 

首先，开发人员首先需要将经过测试后的代码合并到主干，通常是 master 分支，并生成版本号，然后给最新的 commit 打上版本标签。之后，可以将代码 push 到代码托管平台，并触发 CI 流程，CI 流程一般会执行代码扫描、单元测试、编译，最后将构建产物发布到制品库。CI 流程中，可以根据需要添加任意功能。 

接着，进入到发布审批阶段。首先需要申请资源，资源申请周期可能会比较久，所以申请得越早越好，甚至资源申请可以在测试阶段发起。在资源申请阶段，可以申请诸如服务器、MySQL、Redis、Kafka 之类资源。 

资源申请通常是开发人员向运维人员提需求，由运维人员根据需求，在指定的时间前准备好各类资源。如果是物理机通常申请周期会比较久，但当前越来越多的项目选择容器化部署，这可以极大地缩短资源的申请周期。如果在像腾讯云弹性容器这类 Serverless 容器平台上部署业务，甚至可以秒申请资源。所以这里，也建议优先采用容器化部署。 

发布之前需要创建发布计划，里面需要详细描述本次的变更详情，例如变更范围、发布方案、测试结果、验证和回滚方案等。这里需要注意，在创建发布计划时，一定要全面梳理这次变更的影响点。例如，是否有不兼容的变更，是否需要变更配置，是否需要变更数据库等。任何一个遗漏，都可能造成现网故障，影响产品声誉和用户使用。 

接下来，需要创建发布单，在发布单中可以附上发布计划，并根据团队需求填写其它发布内容，发布计划需要跟相关参与者对齐流程、明确职责。发布单最终提交给审批人（通常是技术 leader）对本次发布进行审批，审批通过后，才可以进行部署。 

最后，就可以进入到服务发布阶段，将服务发布到现网。在正式部署的时候，应用需要先部署到预发环境。在预发环境，产品人员、测试人员和研发人员会分别对产品进行验证。 其中，产品人员主要验证产品功能的体验是否流畅，开发和测试人员主要验证产品是否有 Bug。预发环境验证通过，产品才能正式发布到现网。 

这里，强烈建议，编写一些自动化的测试用例，在服务发布到现网之后，对现网服务做一次比较充分的回归测试。通过这个自动化测试，可以以最小的代价，最快速地验证现网功能，从而保障发布质量。 

另外，还要注意，现网可能有多个地域，每个地域发布完成之后都要进行现网验证。 

发布阶段的产出物是正式上线的软件。

#### 运营阶段 

研发流程的最后一个阶段是运营阶段，该阶段主要分为产品运营和运维两个部分。

- 产品运营：通过一系列的运营活动，比如线下的技术沙龙、线上的免费公开课、提高关键词排名或者输出一些技术推广文章等方式，来推高整个产品的知名度，提高产品的用户数量，并提高月活和日活。 
- 运维：由运维工程师负责，核心目标是确保系统稳定的运行，如果系统异常，能够及时发现并修复问题。长期目标是通过技术手段或者流程来完善整个系统架构，减少人力投入、提高运维效率，并提高系统的健壮性和恢复能力。

从上面可以看到，运维属于技术类，运营属于产品类，这二者不要搞混。为了加深理解和记忆，将这些内容，总结在了下面一张图中。

![image-20211109004723363](IAM-document.assets/image-20211109004723363.png)

在运营阶段，研发人员的主要职责就是协助运维解决现网 Bug，优化部署架构。当然，研发人员可能也需要配合运营人员开发一些运营接口，供运营人员使用。 

到这里，业界相对标准的这套研发流程，就学完了。在学习过程中，肯定也发现了，整个研发流程会涉及很多角色，不同角色参与不同的阶段，负责不同的任务。这里再给你额外扩展一个点，就是这些核心角色和分工是啥样的。 

这些扩展内容，放在了一张图和一张表里。这些角色和分工比较好理解，也不需要背下来，只要先有一个大概的印象就可以了。

![image-20211109005006469](IAM-document.assets/image-20211109005006469.png)

具体分工如下表所示。

![image-20211109005045538](IAM-document.assets/image-20211109005045538.png)

### 总结 

在开发 Go 项目时，掌握项目的研发流程很重要。掌握研发流程，会让项目研发对我们更加白盒，并且有利于制定详细的工作任务。 

那么如何设计项目研发流程呢？可以根据需要自行设计。自行设计时有些点是一定要关注的，例如流程需要支持高的发布效率和发布质量，支持快速迭代，流程是合理、 可扩展的，等等。

如果不想自己设计，也可以。介绍了一套相对通用、标准的研发流程， 如果合适可以直接拿来作为自己设计的研发流程。 这套研发流程包含 6 个阶段：需求阶段、设计阶段、开发阶段、测试阶段、发布阶段和运营阶段。这里将这些流程和每个流程的核心点总结在下面一张图中。

![image-20211109005427628](IAM-document.assets/image-20211109005427628.png)

### 课后练习

- 回忆下研发阶段具体包括哪些工作内容，如果觉得这些工作内容满足不了研发阶段的需求，还需要补充什么呢？ 
- 思考、调研下有哪些工具，可以帮助实现整个流程，以及流程中任务的自动化，看下它们是如何提高我们的研发效率的





## 研发流程设计之管理应用的生命周期

如何管理应用生命周期。 

上一讲，介绍了一个相对标准的研发流程，这个研发流程可以确保高效地开发出一个优秀的 Go 项目。

这一讲，再来看下，如何管理 Go 项目，也就是说如何对应用的生命周期进行管理。 

那应用的生命周期管理，怎么理解呢？其实，就是指采用一些好的工具或方法在应用的整个生命周期中对应用进行管理，以提高应用的研发效率和质量。 

那么，如何设计一套优秀的应用生命周期管理手段呢？这就跟研发流程“设计”的思路一样，可以自己设计，也可以采用业界沉淀下来的优秀管理手段。同样地，更建议采用已有的最佳实践，因为重复造轮子、造一个好轮子太难了。

所以，这一讲就一起学习下，业界在不同时期沉淀下来的优秀管理手段，以及对这些管理手段的经验和建议，选到一个最合适的。 

### 应用生命周期管理技术有哪些？ 

那么，有哪些应用生命周期管理技术呢？ 

可以从两个维度来理解应用生命周期管理技术。 

- 第一个维度是演进维度。
  - 应用生命周期，最开始主要是通过研发模式来管理的，按时间线先后出现了瀑布模式、迭代模式、敏捷模式。
  - 接着，为了解决研发模式中的一些痛点出现了另一种管理技术，也就是 CI/CD 技术。随着 CI/CD 技术的成熟，又催生了另一种更高级 的管理技术 DevOps。 
- 第二个维度是管理技术的类别。应用生命周期管理技术可以分为两类：
  - 研发模式，用来确保整个研发流程是高效的。 
  - DevOps，主要通过协调各个部门之间的合作，来提高软件的发布效率和质量。DevOps 中又包含了很多种技术，主要包括 CI/CD 和多种 Ops，例如 AIOps、ChatOps、 GitOps、NoOps 等。其中，CI/CD 技术提高了软件的发布效率和质量，而 Ops 技术则 提高了软件的运维和运营效率。

尽管这些应用生命周期管理技术有很多不同，但是它们彼此支持、相互联系。研发模式专注于开发过程，DevOps 技术里的 CI/CD 专注于流程，Ops 则专注于实战。 

为了帮助理解，总结出了下面这张图供参考。

![image-20211109221017511](IAM-document.assets/image-20211109221017511.png)

这两个维度涉及的管理技术虽然不少，但一共就是那几类。所以，为了能够逻辑清晰地讲解明白这些技术，会从演进维度来展开，也就是按照这样的顺序：研发模式（瀑布模式 -> 迭代模式 -> 敏捷模式） -> CI/CD -> DevOps。 

既然是演进，那这些技术肯定有优劣之分，应该怎么选择呢，一定是选 择后面出现的技术吗？ 

为了解决这个问题，这里，对于研发模式和 DevOps 这两类技术的选择，给出建议：

- 研发模式建议选择敏捷模式，因为它更能胜任互联网时代快速迭代的诉求。 
- DevOps 则要优先确保落地 CI/CD 技术，接着尝试落地 ChatOps 技术，如果有条件可以积极探索 AIOps 和 GitOps。 

接下来，就详细说说这些应用生命周期的管理方法，先来看专注于开发过程的研发模式部分。

### 研发模式 

研发模式主要有三种，演进顺序为瀑布模式 -> 迭代模式 -> 敏捷模式，现在逐一看下。 

#### 瀑布模式 

在早期阶段，软件研发普遍采用的是瀑布模式，熟知的 RHEL、Fedora 等系统就是采用瀑布模式。

瀑布模式按照预先规划好的研发阶段来推进研发进度。比如，按照需求阶段、设计阶段、 开发阶段、测试阶段、发布阶段、运营阶段的顺序串行执行开发任务。每个阶段完美完成之后，才会进入到下一阶段，阶段之间通过文档进行交付。整个过程如下图所示。

![image-20211109221456674](IAM-document.assets/image-20211109221456674.png)

瀑布模式最大的优点是简单。它严格按照研发阶段来推进研发进度，流程清晰，适合按项目交付的应用。 

但它的缺点也很明显，最突出的就是这两个：

- 只有在项目研发的最后阶段才会交付给客户。交付后，如果客户发现问题，变更就会非常困难，代价很大。 
- 研发周期比较长，很难适应互联网时代对产品快速迭代的诉求。

为了解决这两个问题，迭代式研发模式诞生了。 

#### 迭代模式 

迭代模式，是一种与瀑布式模式完全相反的开发过程：研发任务被切分为一系列轮次，每一个轮次都是一个迭代，每一次迭代都是一个从设计到实现的完整过程。它不要求每一个阶段的任务都做到最完美，而是先把主要功能搭建起来，然后再通过客户的反馈信息不断完善。

迭代开发可以帮助产品改进和把控进度，它的灵活性极大地提升了适应需求变化的能力， 克服了高风险、难变更、复用性低的特点。 

但是，迭代模式的问题在于比较专注于开发过程，很少从项目管理的视角去加速和优化项目开发过程。接下来要讲的敏捷模式，就弥补了这个缺点。 

#### 敏捷模式 

敏捷模式把一个大的需求分成多个、可分阶段完成的小迭代，每个迭代交付的都是一个可使用的软件。在开发过程中，软件要一直处于可使用状态。 

敏捷模式中具有代表性的开发模式，是 Scrum 开发模型。

在敏捷模式中，会把一个大的需求拆分成很多小的迭代，这意味着开发过程中会有很多个开发、构建、测试、发布和部署的流程。这种高频度的操作会给研发、运维和测试人员带来很大的工作量，降低了工作效率。为了解决这个问题，CI/CD 技术诞生了。

### CI/CD：自动化构建和部署应用 

CI/CD 技术通过自动化的手段，来快速执行代码检查、测试、构建、部署等任务，从而提高研发效率，解决敏捷模式带来的弊端。 

CI/CD 包含了 3 个核心概念。

- CI：Continuous Integration，持续集成。 
- CD：Continuous Delivery，持续交付。 
- CD：Continuous Deployment，持续部署。

CI 容易理解，但两个 CD 很多开发者区分不开。

首先是持续集成。

它的含义为：频繁地（一天多次）将开发者的代码合并到主干上。它的流程为：在开发人员完成代码开发，并 push 到 Git 仓库后，CI 工具可以立即对代码进行扫描、（单元）测试和构建，并将结果反馈给开发者。持续集成通过后，会将代码合并到主干。

CI 流程可以使应用软件的问题在开发阶段就暴露出来，这会让开发人员交付代码时更有信心。因为 CI 流程内容比较多，而且执行比较频繁，所以 CI 流程需要有自动化工具来支撑。 

其次是持续交付。

它指的是一种能够使软件在较短的循环中可靠发布的软件方法。 

持续交付在持续集成的基础上，将构建后的产物自动部署在目标环境中。这里的目标环境，可以是测试环境、预发环境或者现网环境。 

通常来说，持续部署可以自动地将服务部署到测试环境或者预发环境。因为部署到现网环境存在一定的风险，所以如果部署到现网环境，需要手工操作。手工操作的好处是，可以使相关人员评估发布风险，确保发布的正确性。 

最后是持续部署。

持续部署在持续交付的基础上，将经过充分测试的代码自动部署到生产环境，整个流程不再需要相关人员的审核。持续部署强调的是自动化部署，是交付的最高阶段。 

可以借助下面这张图，来了解持续集成、持续交付、持续部署的关系。

![image-20211109222443685](IAM-document.assets/image-20211109222443685.png)

持续集成、持续交付和持续部署强调的是持续性，也就是能够支持频繁的集成、交付和部署，这离不开自动化工具的支持，离开了这些工具，CI/CD 就不再具有可实施性。

持续集成的核心点在代码，持续交付的核心点在可交付的产物，持续部署的核心点在自动部署。 

### DevOps：研发运维一体化 

CI/CD 技术的成熟，加速了 DevOps 这种应用生命周期管理技术的成熟和落地。

DevOps（Development 和 Operations 的组合）是一组过程、方法与系统的统称，用于促进开发（应用程序 / 软件工程）、技术运营和质量保障（QA）部门之间的沟通、协作与整合。这 3 个部门的相互协作，可以提高软件质量、快速发布软件。如下图所示：

![image-20211109222825386](IAM-document.assets/image-20211109222825386.png)

要实现 DevOps，需要一些工具或者流程的支持，CI/CD 可以很好地支持 DevOps 这种软件开发模式，如果没有 CI/CD 自动化的工具和流程，DevOps 就是没有意义的，CI/CD 使 得 DevOps 变得可行。 

听到这里是不是有些晕？可能想问，DevOps 跟 CI/CD 到底是啥区别呢？其实，这也是困扰很多开发者的问题。可以这么理解：DevOps ！= CI/CD。DevOps 是一组过程、方法和系统的统称，而 CI/CD 只是一种软件构建和发布的技术。

DevOps 技术之前一直有，但是落地不好，因为没有一个好的工具来实现 DevOps 的理念。但是随着容器、CI/CD 技术的诞生和成熟，DevOps 变得更加容易落地。也就是说， 这几年越来越多的人采用 DevOps 手段来提高研发效能。 

随着技术的发展，目前已经诞生了很多 Ops 手段，来实现运维和运营的高度自动化。下 面，就来看看 DevOps 中的四个 Ops 手段：AIOps、ChatOps、GitOps、NoOps。 

#### AIOps：智能运维 

在 2016 年，Gartner 提出利用 AI 技术的新一代 IT 运维，即 AIOps（智能运维）。通过 AI 手段，来智能化地运维 IT 系统。AIOps 通过搜集海量的运维数据，并利用机器学习算法，智能地定位并修复故障。 

也就是说，AIOps 在自动化的基础上，增加了智能化，从而进一步推动了 IT 运维自动化， 减少了人力成本。

随着 IT 基础设施规模和复杂度的倍数增长，企业应用规模、数量的指数级增长，传统的人工 / 自动化运维，已经无法胜任愈加沉重的运维工作，而 AIOps 提供了一个解决方案。

在腾讯、阿里等大厂很多团队已经在尝试和使用 AIOps，并享受到了 AIOps 带来的红利。例如，故障告警更加灵敏、准确，一些常见的故障，可以自动修复，无须运维人员介入等。

#### ChatOps：聊着天就把事情给办了 

随着企业微信、钉钉等企业内通讯工具的兴起，最近几年出现了一个新的概念 ChatOps。 

简单来说，ChatOps 就是在一个聊天工具中，发送一条命令给 ChatBot 机器人，然后 ChatBot 会执行预定义的操作。这些操作可以是执行某个工具、调用某个接口等，并返回执行结果。 

这种新型智能工作方式的优势是什么呢？它可以利用 ChatBot 机器人让团队成员和各项辅助工具连接在一起，以沟通驱动的方式完成工作。ChatOps 可以解决人与人、人与工具、 工具与工具之间的信息孤岛，从而提高协作体验和工作效率。 ChatOps 的工作流程如下图所示（网图）：

![image-20211109223514170](IAM-document.assets/image-20211109223514170.png)

开发 / 运维 / 测试人员通过 @聊天窗口中的机器人 Bot 来触发任务，机器人后端会通过 API 接口调用等方式对接不同的系统，完成不同的任务，例如持续集成、测试、发布等工 作。

机器人可以是自己研发的，也可以是开源的。目前，业界有很多流行的机器人可供选择，常用的有 Hubot、Lita、Errbot、StackStorm 等。 

使用 ChatOps 可以带来以下几点好处。

- 友好、便捷：所有的操作均在同一个聊天界面中，通过 @机器人以聊天的方式发送命令，免去了打开不同系统，执行不同操作的繁琐操作，方式更加友好和便捷。 
- 信息透明：在同一个聊天界面中的所有同事都能够看到其他同事发送的命令，以及命令 执行的结果，可以消除沟通壁垒，工作历史有迹可循，团队合作更加顺畅。 
- 移动友好：可以在移动端向机器人发送命令、执行任务，让移动办公变为可能。 
- DevOps 文化打造：通过与机器人对话，可以降低项目开发中，各参与人员的理解和使用成本，从而使 DevOps 更容易落地和推广。

#### GitOps： 一种实现云原生的持续交付模型 

GitOps 是一种持续交付的方式。它的核心思想是将应用系统的声明性基础架构（YAML） 和应用程序存放在 Git 版本库中。将 Git 作为交付流水线的核心，每个开发人员都可以提交拉取请求（Pull Request），并使用 Gi t 来加速和简化 Kubernetes 的应用程序部署和运维任务。

通过 Git 这样的工具，开发人员可以将精力聚焦在功能开发，而不是软件运维上，以此提高软件的开发效率和迭代速度。 

使用 GitOps 可以带来很多优点，其中最核心的是：当使用 Git 变更代码时，GitOps 可以自动将这些变更应用到程序的基础架构上。因为整个流程都是自动化的，所以部署时间更短；又因为 Git 代码是可追溯的，所以部署的应用也能够稳定且可重现地回滚。 

可以从概念和流程上来理解 GitOps，它有 3 个关键概念。

- 声明性容器编排：通过 Kubernetes YAML 格式的资源定义文件，来定义如何部署应用。 
- 不可变基础设施：基础设施中的每个组件都可以自动的部署，组件在部署完成后，不能发生变更。如果需要变更，则需要重新部署一个新的组件。例如，Kubernetes 中的 Pod 就是一个不可变基础设施。 
- 连续同步：不断地查看 Git 存储库，将任何状态更改反映到 Kubernetes 集群中。

<img src="IAM-document.assets/image-20211109224350340.png" alt="image-20211109224350340"/>

GitOps 的工作流程如下： 

首先，开发人员开发完代码后推送到 Git 仓库，触发 CI 流程，CI 流程通过编译构建出 Docker 镜像，并将镜像 push 到 Docker 镜像仓库中。Push 动作会触发一个 push 事件，通过 webhook 的形式通知到 Config Updater 服务，Config Updater 服务会从 webhook 请求中获取最新 push 的镜像名，并更新 Git 仓库中的 Kubernetes YAML 文 件。

然后，GitOps 的 Deploy Operator 服务，检测到 YAML 文件的变动，会重新从 Git 仓库 中提取变更的文件，并将镜像部署到 Kubernetes 集群中。Config Updater 和 Deploy Operator 两个组件需要开发人员设计开发。 

#### NoOps：无运维 

NoOps 即无运维，完全自动化的运维。在 NoOps 中不再需要开发人员、运营运维人员的协同，把微服务、低代码、无服务全都结合了起来，开发者在软件生命周期中只需要聚焦业务开发即可，所有的维护都交由云厂商来完成。 

毫无疑问，NoOps 是运维的终极形态，它像 DevOps 一样，更多的是一种理 念，需要很多的技术和手段来支撑。当前整个运维技术的发展，也是朝着 NoOps 的方向去演进的，例如 GitOps、AIOps 可以使我们尽可能减少运维，Serverless 技术甚至可以使我们免运维。相信未来 NoOps 会像现在的 Serverless 一样，成为一种流行的、可落地的理念。

### 如何选择合适的应用生命周期管理技术？ 

在实际开发中， 如何选择适合自己的呢？可以从这么几个方面考虑。 

首先，根据团队、项目选择一个合适的研发模式。如果项目比较大，需求变更频繁、要求 快速迭代，建议选择敏捷开发模式。敏捷开发模式，也是很多大公司选择的研发模式，在 互联网时代很受欢迎。 

接着，要建立自己的 CI/CD 流程。任何变更代码在合并到 master 分支时，一定要通过 CI/CD 的流程的验证。建议 在 CI/CD 流程中设置质量红线，确保合并代码的质量。 

接着，除了建立 CI/CD 系统，还建议将 ChatOps 带入工作中，尽可能地将可以自动化的工作实现自动化，并通过 ChatOps 来触发自动化流程。随着企业微信、钉钉等企业聊天软件成熟和发展，ChatOps 变得流行和完善。 

最后，GitOps、AIOps 可以将部署和运维自动化做到极致，在团队有人力的情况下，值得探索。

大厂是如何管理应用生命周期的？ 

大厂普遍采用敏捷开发的模式，来适应互联网对应用快速迭代的诉求。例如，腾讯的 TAPD、Coding的 Scrum 敏捷管理就是一个敏捷开发平台。CI/CD 强制落地， ChatOps 已经广泛使用，AIOps 也有很多落地案例，GitOps 目前还在探索阶段， NoOps 还处在理论阶段。 

### 总结 

从技术演进的维度介绍了应用生命周期管理技术，这些技术可以提高应用的研 发效率和质量。 

应用生命周期管理最开始是通过研发模式来管理的。在研发模式中，按时间线分别介绍了瀑布模式、迭代模式和敏捷模式，其中的敏捷模式适应了互联网时代对应用快速迭代的诉求，所以用得越来越多。 

在敏捷模式中，需要频繁构建和发布应用，这就给开发人员带来了额外的工作 量，为了解决这个问题，出现了 CI/CD 技术。CI/CD 可以将代码的检查、测试、构建和部署等工作自动化，不仅提高了研发效率，还从一定程度上保障了代码的质量。另外，CI/CD 技术使得 DevOps 变得可行，当前越来越多的团队采用 DevOps 来管理应用的生命周期。 另

外，介绍了几个大家容易搞混的概念。

- 持续交付和持续部署。二者都是持续地部署应用，但是持续部署整个过程是自动化的， 而持续交付中，应用在发布到现网前需要人工审批是否允许发布。 
- CI/CD 和 DevOps。DevOps 是一组过程、方法与系统的统称，其中也包含了 CI/CD 技术。而 CI/CD 是一种自动化的技术，DevOps 理念的落地需要 CI/CD 技术的支持。

最后，关于如何管理应用的生命周期，给出了一些建议：研发模式建议选择敏捷模式， 因为它更能胜任互联网时代快速迭代的诉求。DevOps 则要优先确保落地 CI/CD 技术，接着尝试落地 ChatOps 技术，如果有条件可以积极探索 AIOps 和 GitOps。 

### 课后练习

- 学习并使用 GitHub Actions，通过 Github Actions 完成提交代码后自动进行静态代码检查的任务。 
- 尝试添加一个能够每天自动打印“hello world”的企业微信机器人，并思考下，哪些自动化工作可以通过该机器人来实现。



## 设计方法之优雅的开发Go项目

如何写出优雅的 Go 项目。 

Go 语言简单易学，对于大部分开发者来说，编写可运行的代码并不是一件难事，但如果想真正成为 Go 编程高手，需要花很多精力去研究 Go 的编程哲学。 

在 Go 开发生涯中，见过各种各样的代码问题，例如：代码不规范，难以阅读；函数共享性差，代码重复率高；不是面向接口编程，代码扩展性差，代码不可测；代码质量低下。

究其原因，是因为这些代码的开发者很少花时间去认真研究如何开发一个优雅的 Go 项目，更多时间是埋头在需求开发中。 

如果遇到过以上问题，那么是时候花点时间来研究下如何开发一个优雅的 Go 项目了。只有这样，才能区别于绝大部分的 Go 开发者，从而在职场上建立自己的核心竞争力，并最终脱颖而出。 

其实，之前所学的各种规范设计，也都是为了写出一个优雅的 Go 项目。在这一讲， 又补充了一些内容，从而形成了一套“写出优雅 Go 项目”的方法论。

### 如何写出优雅的 Go 项目？ 

那么，如何写出一个优雅的 Go 项目呢？在回答这个问题之前，先来看另外两个问题：

1. 为什么是 Go 项目，而不是 Go 应用？ 
2. 一个优雅的 Go 项目具有哪些特点？

先来看第一个问题。Go 项目是一个偏工程化的概念，不仅包含了 Go 应用，还包含了项目管理和项目文档：

![image-20211110221136606](IAM-document.assets/image-20211110221136606.png)

这就来到了第二个问题，一个优雅的 Go 项目，不仅要求 Go 应用是优雅的，还要确保项目管理和文档也是优雅的。这样，根据前面学到的 Go 设计规范， 很容易就能总结出一个优雅的 Go 应该具备的特点：

- 符合 Go 编码规范和最佳实践； 
- 易阅读、易理解，易维护； 
- 易测试、易扩展； 
- 代码质量高。

解决了这两个问题，回到这一讲的核心问题：如何写出优雅的 Go 项目？ 

写出一个优雅的 Go 项目，就是用“最佳实践”的方式去实现 Go 项目中的 Go 应用、项目管理和项目文档。具体来说，就是编写高质量的 Go 应用、高效管理项目、编写高质量的项目文档。

为了协助理解，将这些逻辑绘制成了下面一张图。

![image-20211110221422191](IAM-document.assets/image-20211110221422191.png)

接下来，就看看如何根据前面学习的 Go 项目设计规范，实现一个优雅的 Go 项 目。先从编写高质量的 Go 应用看起。 

### 编写高质量的 Go 应用 

基于研发经验，要编写一个高质量的 Go 应用，其实可以归纳为 5 个方面：代码结构、代码规范、代码质量、编程哲学和软件设计方法，见下图。 

![image-20211110221549253](IAM-document.assets/image-20211110221549253.png)

#### 代码结构 

为什么先说代码结构呢？因为组织合理的代码结构是一个项目的门面。可以通过两个手段来组织代码结构。 

第一个手段是，组织一个好的目录结构。关于如何组合一个好的目录结构，可以回顾 《目录结构设计》 的内容。 

第二个手段是，选择一个好的模块拆分方法。做好模块拆分，可以使项目内模块职责分明，做到低耦合高内聚。 

那么 Go 项目开发中，如何拆分模块呢？目前业界有两种拆分方法，分别是按层拆分和按功能拆分。 

##### 按层拆分

首先，看下按层拆分，最典型的是 MVC 架构中的模块拆分方式。在 MVC 架构中， 将服务中的不同组件按访问顺序，拆分成了 Model、View 和 Controller 三层。

![image-20211110221846218](IAM-document.assets/image-20211110221846218.png)

每层完成不同的功能：

- View（视图）是提供给用户的操作界面，用来处理数据的显示。
- Controller（控制器），负责根据用户从 View 层输入的指令，选取 Model 层中的数据，然后对其进行相应的操作，产生最终结果。 
- Model（模型），是应用程序中用于处理数据逻辑的部分。

看一个典型的按层拆分的目录结构：

```sh
$ tree --noreport -L 2 layers
layers
├── controllers
│ 	├── billing
│ 	├── order
│ 	└── user
├── models
│ 	├── billing.go
│ 	├── order.go
│ 	└── user.go
└── views
		└── layouts
```

在 Go 项目中，按层拆分会带来很多问题。最大的问题是循环引用：相同功能可能在不同层被使用到，而这些功能又分散在不同的层中，很容易造成循环引用。 

所以，只要大概知道按层拆分是什么意思就够了，在 Go 项目中建议使用的是按功能拆分的方法，这也是 Go 项目中最常见的拆分方法。 

##### 按功能拆分

那什么是按功能拆分呢？看一个例子。

比如，一个订单系统，可以根据不同功能将其拆分成用户（user）、订单（order）和计费（billing）3 个模块，每一个模块提供独立的功能，功能更单一：

![image-20211110222547522](IAM-document.assets/image-20211110222547522.png)

 下面是该订单系统的代码目录结构：

```sh
$ tree pkg
$ tree --noreport -L 2 pkg
pkg
├── billing
├── order
│ 	└── order.go
└── user
```

相较于按层拆分，按功能拆分模块带来的好处也很好理解：

- 不同模块，功能单一，可以实现高内聚低耦合的设计哲学。 
- 因为所有的功能只需要实现一次，引用逻辑清晰，会大大减少出现循环引用的概率。

所以，有很多优秀的 Go 项目采用的都是按功能拆分的模块拆分方式，例如 Kubernetes、 Docker、Helm、Prometheus 等。 

除了组织合理的代码结构这种方式外，编写高质量 Go 应用的另外一个行之有效的方法， 是遵循 Go 语言代码规范来编写代码。

#### 代码规范

要遵循哪些代码规范来编写 Go 应用呢？其实就两类：编码规范和最佳实践。 

##### 编码规范

首先，代码要符合 Go 编码规范，这是最容易实现的途径。Go 社区有很多这类规范可供参考，其中，比较受欢迎的是 Uber Go 语言编码规范。 

阅读这些规范确实有用，也确实花时间、花精力。所以，在参考了已有的很多规范后， 结合自己写 Go 代码的经验，整理了一篇 Go 编码规范作为加餐。 

有了可以参考的编码规范之后，需要扩展到团队、部门甚至公司层面。只有大家一起参与、遵守，规范才会变得有意义。其实，大家都清楚，要开发者靠自觉来遵守所有的编码规范，不是一件容易的事儿。这时候，可以使用静态代码检查工具，来约束开发者的行为。 

有了静态代码检查工具后，不仅可以确保开发者写出的每一行代码都是符合 Go 编码规范的，还可以将静态代码检查集成到 CI/CD 流程中。这样，在代码提交后自动地检查代码， 就保证了只有符合编码规范的代码，才会被合入主干。

Go 语言的静态代码检查工具有很多，目前用的最多的是 golangci-lint，这也是极力推荐使用的一个工具。关于这个工具的使用。

##### 最佳实践

除了遵循编码规范，要想成为 Go 编程高手，还得学习并遵循一些最佳实践。“最佳实践”是社区经过多年探索沉淀下来的、符合 Go 语言特色的经验和共识，可以帮助大家开发出一个高质量的代码。 

这里推荐几篇介绍 Go 语言最佳实践的文章，供参考：

- Effective Go：高效 Go 编程，由 Golang 官方编写，里面包含了编写 Go 代码的一些建议，也可以理解为最佳实践。 
- Go Code Review Comments：Golang 官方编写的 Go 最佳实践，作为 Effective Go 的补充。 
- Style guideline for Go packages：包含了如何组织 Go 包、如何命名 Go 包、如何写 Go 包文档的一些建议。

#### 代码质量 

有了组织合理的代码结构、符合 Go 语言代码规范的 Go 应用代码之后，还需要通过一些手段来确保开发出的是一个高质量的代码，这可以通过单元测试和 Code Review 来实现。 

##### 单元测试

单元测试非常重要。

开发完一段代码后，第一个执行的测试就是单元测试。它可以保证代码是符合预期的，一些异常变动能够被及时感知到。进行单元测试，不仅需要编写单元测试用例，还需要确保代码是可测试的，以及具有一个高的单元测试覆盖率。 

接下来，就来介绍下如何编写一个可测试的代码。 

如果要对函数 A 进行测试，并且 A 中的所有代码均能够在单元测试环境下按预期被执行，那么函数 A 的代码块就是可测试的。来看下一般的单元测试环境有什么特点：

- 可能无法连接数据库。 
- 可能无法访问第三方服务。

如果函数 A 依赖数据库连接、第三方服务，那么在单元测试环境下执行单元测试就会失败，函数就没法测试，函数是不可测的。 

解决方法也很简单：将依赖的数据库、第三方服务等抽象成接口，在被测代码中调用接口的方法，在测试时传入 mock 类型，从而将数据库、第三方服务等依赖从具体的被测函数中解耦出去。如下图所示：

![image-20211110223955529](IAM-document.assets/image-20211110223955529.png)

为了提高代码的可测性，降低单元测试的复杂度，对 function 和 mock 的要求是：

- 要尽可能减少 function 中的依赖，让 function 只依赖必要的模块。编写一个功能单一、职责分明的函数，会有利于减少依赖。 
- 依赖模块应该是易 Mock 的。

为了协助理解，先来看一段不可测试的代码：

```go
package main

import "google.golang.org/grpc"

type Post struct {
	Name    string
	Address string
}

func ListPosts(client *grpc.ClientConn) ([]*Post, error) {
	return client.ListPosts()
}
```

这段代码中的 ListPosts 函数是不可测试的。因为 ListPosts 函数中调用了 client.ListPosts()方法，该方法依赖于一个 gRPC 连接。而在做单元测试时， 可能因为没有配置 gRPC 服务的地址、网络隔离等原因，导致没法建立 gRPC 连接，从而导致 ListPosts 函数执行失败。 

下面，把这段代码改成可测试的，如下：

```go
package main

type Post struct {
	Name    string
	Address string
}
type Service interface {
	ListPosts() ([]*Post, error)
}

func ListPosts(svc Service) ([]*Post, error) {
	return svc.ListPosts()
}
```

上面代码中，ListPosts 函数入参为 Service 接口类型，只要传入一个实现了 Service 接口类型的实例，ListPosts 函数即可成功运行。因此，可以在单元测试中可以实现一个不依赖任何第三方服务的 fake 实例，并传给 ListPosts。

上述可测代码的单元测试代码如下：

```go
package main

import "testing"

type fakeService struct {
}

func NewFakeService() Service {
	return &fakeService{}
}
func (s *fakeService) ListPosts() ([]*Post, error) {
	posts := make([]*Post, 0)
	posts = append(posts, &Post{
		Name:    "colin",
		Address: "Shenzhen",
	})
	posts = append(posts, &Post{
		Name:    "alex",
		Address: "Beijing",
	})
	return posts, nil
}
func TestListPosts(t *testing.T) {
	fake := NewFakeService()
	if _, err := ListPosts(fake); err != nil {
		t.Fatal("list posts failed")
	}
}
```

当代码可测之后，就可以借助一些工具来 Mock 需要的接口了。常用的 Mock 工具，有这么几个：

- golang/mock，是官方提供的 Mock 框架。它实现了基于 interface 的 Mock 功 能，能够与 Golang 内置的 testing 包做很好的集成，是最常用的 Mock 工具。 golang/mock 提供了 mockgen 工具用来生成 interface 对应的 Mock 源文件。 
- sqlmock，可以用来模拟数据库连接。数据库是项目中比较常见的依赖，在遇到数据库依赖时都可以用它。 
- httpmock，可以用来 Mock HTTP 请求。 
- bouk/monkey，猴子补丁，能够通过替换函数指针的方式来修改任意函数的实现。 如果 golang/mock、sqlmock 和 httpmock 这几种方法都不能满足需求，可以尝试通过猴子补丁的方式来 Mock 依赖。可以这么说，猴子补丁提供了单元测试 Mock 依赖的最终解决方案。

##### 单元测试覆盖率

接下来，再一起看看如何提高单元测试覆盖率。 

当编写了可测试的代码之后，接下来就需要编写足够的测试用例，用来提高项目的单元测试覆盖率。这里有以下两个建议供参考：

- 使用 gotests 工具自动生成单元测试代码，减少编写单元测试用例的工作量，从重复的劳动中解放出来。 

- 定期检查单元测试覆盖率。可以通过以下方法来检查：

  - ```sh
    $ go test -race -cover -coverprofile=./coverage.out -timeout=10m -short -v ./...
    $ go tool cover -func ./coverage.out
    ```

  - 执行结果如下：

    - ![image-20211110231130517](IAM-document.assets/image-20211110231130517.png)

在提高项目的单元测试覆盖率时，可以先提高单元测试覆盖率低的函数，之后再检查项目的单元测试覆盖率；如果项目的单元测试覆盖率仍然低于期望的值，可以再次提高单元测试覆盖率低的函数的覆盖率，然后再检查。以此循环，最终将项目的单元测试覆盖率优化到预期的值为止。 

这里要注意，对于一些可能经常会变动的函数单元测试，覆盖率要达到 100%。 

##### Code Review(CR)

说完了单元测试，再看看如何通过 Code Review 来保证代码质量。 

Code Review 可以提高代码质量、交叉排查缺陷，并且促进团队内知识共享，是保障代码质量非常有效的手段。在项目开发中，一定要建立一套持久可行的 Code Review 机制。 

但在研发生涯中，发现很多团队没有建立有效的 Code Review 机制。这些团队都认可 Code Review 机制带来的好处，但是因为流程难以遵守，慢慢地 Code Review 就变成了 形式主义，最终不了了之。

其实，建立 Code Review 机制很简单，主要有 3 点：

- 首先，确保使用的代码托管平台有 Code Review 的功能。比如，GitHub、GitLab 这类代码托管平台都具备这种能力。
-  接着，建立一套 Code Review 规范，规定如何进行 Code Review。 
- 最后，也是最重要的，每次代码变更，相关开发人员都要去落实 Code Review 机制， 并形成习惯，直到最后形成团队文化。

到这里可以小结一下：组织一个合理的代码结构、编写符合 Go 代码规范的代码、保证代码质量，都是编写高质量 Go 代码的外功。

那内功是什么呢？就是编程哲学和软件设计方法。 

#### 编程哲学 

那编程哲学是什么意思呢？编程哲学，其实就是要编写符合 Go 语言设计哲学的代码。Go 语言有很多设计哲学，对代码质量影响比较大的，有两个：面向接口编程和面向“对象”编程。 

##### 面向接口编程

先来看下面向接口编程。

Go 接口是一组方法的集合。任何类型，只要实现了该接口中的方法集，那么就属于这个类型，也称为实现了该接口。 

接口的作用，其实就是为不同层级的模块提供一个定义好的中间层。这样，上游不再需要依赖下游的具体实现，充分地对上下游进行了解耦。很多流行的 Go 设计模式，就是通过面向接口编程的思想来实现的。 

看一个面向接口编程的例子。下面这段代码定义了一个Bird接口，Canary 和 Crow 类型均实现了Bird接口。

```go
package main

import "fmt"

// Bird 定义了一个鸟类
type Bird interface {
	Fly()
	Type() string
}

// Canary 鸟类：金丝雀
type Canary struct {
	Name string
}

func (c *Canary) Fly() {
	fmt.Printf("我是%s，用黄色的翅膀飞\n", c.Name)
}
func (c *Canary) Type() string {
	return c.Name
}

// Crow 鸟类：乌鸦
type Crow struct {
	Name string
}

func (c *Crow) Fly() {
	fmt.Printf("我是%s，用黑色的翅膀飞\n", c.Name)
}
func (c *Crow) Type() string {
	return c.Name
}

// LetItFly 让鸟类飞一下
func LetItFly(bird Bird) {
	fmt.Printf("Let %s Fly!\n", bird.Type())
	bird.Fly()
}

func main() {
	LetItFly(&Canary{"金丝雀"})
	LetItFly(&Crow{"乌鸦"})
}
```

这段代码中，因为 Crow 和 Canary 都实现了 Bird 接口声明的 Fly、Type 方法，所以可以说 Crow、Canary 实现了 Bird 接口，属于 Bird 类型。在函数调用时，可以传入 Bird 类型，并在函数内部调用 Bird 接口提供的方法，以此来解耦 Bird 的具体实现。 

总结下使用接口的好处：

- 代码扩展性更强了。例如，同样的 Bird，可以有不同的实现。在开发中用的更多的是， 将数据库的 CURD 操作抽象成接口，从而可以实现同一份代码对接不同数据库的目的。 
- 可以解耦上下游的实现。例如，LetItFly 不用关注 Bird 是如何 Fly 的，只需要调用 Bird 提供的方法即可。 
- 提高了代码的可测性。因为接口可以解耦上下游实现，在单元测试需要依赖第三方系统 / 数据库的代码时，可以利用接口将具体实现解耦，实现 fake 类型。 
- 代码更健壮、更稳定了。例如，如果要更改 Fly 的方式，只需要更改相关类型的 Fly 方法即可，完全影响不到 LetItFly 函数。

所以，在 Go 项目开发中，一定要多思考，哪些可能有多种实现的地方，要考虑使用接口。 

##### 面向对象编程

接下来，再来看下面向“对象”编程。 

面向对象编程（OOP）有很多优点，可以使代码变得易维护、易扩展，并能提高开发效率等，所以一个高质量的 Go 应用在需要时，也应该采用面向对象的方法去编程。

那什么叫“在需要时”呢？就是在开发代码时，如果一个功能可以通过接近于日常生活和自然的思考方式来实现，这时候就应该考虑使用面向对象的编程方法。

Go 语言不支持面向对象编程，但是却可以通过一些语言级的特性来实现类似的效果。 

面向对象编程中，有几个核心特性：类、实例、抽象，封装、继承、多态、构造函数、析构函数、方法重载、this 指针。在 Go 中可以通过以下几个方式来实现类似的效果：

- 类、抽象、封装通过结构体来实现。 
- 实例通过结构体变量来实现。 
- 继承通过组合来实现。这里解释下什么叫组合：一个结构体嵌到另一个结构体，称作组合。例如一个结构体包含了一个匿名结构体，就说这个结构体组合了该匿名结构体。 
- 多态通过接口来实现。

至于构造函数、析构函数、方法重载和 this 指针等，Go 为了保持语言的简洁性去掉了这些特性。 

Go 中面向对象编程方法，见下图：

![image-20211110233502125](IAM-document.assets/image-20211110233502125.png)

通过一个示例，来具体看下 Go 是如何实现面向对象编程中的类、抽象、封装、继承和多态的。代码如下：

```go
package main

import "fmt"

// Bird 基类：Bird
type Bird struct {
	Type string
}

// Class 鸟的类别
func (bird *Bird) Class() string {
	return bird.Type
}

// Birds 定义了一个鸟类
type Birds interface {
	Name() string
	Class() string
}

// Canary 鸟类：金丝雀
type Canary struct {
	Bird
	name string
}
func (c *Canary) Name() string {
	return c.name
}

// Crow 鸟类：乌鸦
type Crow struct {
	Bird
	name string
}
func (c *Crow) Name() string {
	return c.name
}

func NewCrow(name string) *Crow {
	return &Crow{
		Bird: Bird{
			Type: "Crow",
		},
		name: name,
	}
}
func NewCanary(name string) *Canary {
	return &Canary{
		Bird: Bird{
			Type: "Canary",
		},
		name: name,
	}
}
func BirdInfo(birds Birds) {
	fmt.Printf("I'm %s, I belong to %s bird class!\n", birds.Name(), birds.Class())
}

func main() {
	canary := NewCanary("CanaryA")
	crow := NewCrow("CrowA")
	BirdInfo(canary)
	BirdInfo(crow)
}
```

将上述代码保存在 oop.go 文件中，执行以下代码输出如下：

```sh
$ go run oop.go
I'm CanaryA, I belong to Canary bird class!
I'm CrowA, I belong to Crow bird class!
```

在上面的例子中，分别通过 Canary 和 Crow 结构体定义了金丝雀和乌鸦两种类别的鸟， 其中分别封装了 name 属性和 Name 方法。也就是说通过结构体实现了类，该类抽象了鸟类，并封装了该鸟类的属性和方法。 

在 Canary 和 Crow 结构体中，都有一个 Bird 匿名字段，Bird 字段为 Canary 和 Crow 类的父类，Canary 和 Crow 继承了 Bird 类的 Class 属性和方法。也就是说通过匿名字段实现了继承。 

在 main 函数中，通过 NewCanary 创建了 Canary 鸟类实例，并将其传给 BirdInfo 函数。也就是说通过结构体变量实现实例。 

在 BirdInfo 函数中，将 Birds 接口类型作为参数传入，并在函数中调用了 birds.Name， birds.Class 方法，这两个方法会根据 birds 类别的不同而返回不同的名字和类别，也就是说通过接口实现了多态。

#### 软件设计方法 

接下来，继续学习编写高质量 Go 代码的第二项内功，也就是让编写的代码遵循一些业界沉淀下来的，优秀的软件设计方法。 

优秀的软件设计方法有很多，其中有两类方法对代码质量的提升特别有帮助，分别是设计模式（Design pattern）和 SOLID 原则。 

设计模式可以理解为业界针对一些特定的场景总结出来的最佳实现方式。它的特点是解决的场景比较具体，实施起来会比较简单；而 SOLID 原则更侧重设计原则，需要彻底理解，并在编写代码时多思考和落地。

关于设计模式和 SOLID 原则，会在设计模式（下一节）中，学习 Go 项目常用的设计模式；至于 SOLID 原则，会简单告诉这个原则是啥。 

##### 设计模式（Design pattern）

先了解下有哪些设计模式。 

在软件领域，沉淀了一些比较优秀的设计模式，其中最受欢迎的是 GOF 设计模式。GOF 设计模式中包含了 3 大类（创建型模式、结构型模式、行为型模式），共 25 种经典的、 可以解决常见软件设计问题的设计方案。这 25 种设计方案同样也适用于 Go 语言开发的项目。 

将这 25 种设计模式总结成了一张图，对于一些在 Go 项目开发中常用的设计模式，会在后续详细介绍。

![image-20211110235253996](IAM-document.assets/image-20211110235253996.png)

##### SOLID 原则

如果说设计模式解决的是具体的场景，那么 SOLID 原则就是设计应用代码时的指导方针。 

SOLID 原则，是由 罗伯特·C·马丁 在 21 世纪早期引入的，包括了面向对象编程和面向对象设计的五个基本原则： 

![image-20211110235450674](IAM-document.assets/image-20211110235450674.png)

遵循 SOLID 原则可以确保设计的代码是易维护、易扩展、易阅读的。SOLID 原则同样也适用于 Go 程序设计。 

如果需要更详细地了解 SOLID 原则，可以参考下 SOLID 原则介绍 这篇文章。 

到这里，就学完了“编写高质量的 Go 应用”这部分内容。接下来，再来学习下 如何高效管理 Go 项目，以及如何编写高质量的项目文档。这里面的大部分内容，之前都有学习过，因为它们是“如何写出优雅的 Go 项目”的重要组成部分，所以，仍然会简单介绍下它们。

### 高效管理项目

一个优雅的 Go 项目，还需要具备高效的项目管理特性。那么如何高效管理项目呢？ 

不同团队、不同项目会采用不同的方法来管理项目，比较重要的有 3 点，分别是制定一个高效的开发流程、使用 Makefile 管理项目和将项目管理自动化。

可以通过自动生成代码、借助工具、对接 CI/CD 系统等方法来将项目管理自动化。具体见下图： 

![image-20211111000003526](IAM-document.assets/image-20211111000003526.png)

#### 高效的开发流程 

高效管理项目的第一步，就是要有一个高效的开发流程，这可以提高开发效率、减少软件维护成本。可以回想一下研发流程设计之工业标准的知识。

#### 使用 Makefile 管理项目 

为了更好地管理项目，除了一个高效的开发流程之外，使用 Makefile 也很重要。Makefile 可以将项目管理的工作通过 Makefile 依赖的方式实现自动化，除了可以提高管理效率之外，还能够减少人为操作带来的失误，并统一操作方式，使项目更加规范。 

IAM 项目的所有操作均是通过 Makefile 来完成的，具体 Makefile 完成了如下操作：

```makefile
build 				Build source code for host platform.
build.multiarch     Build source code for multiple platforms. See option PLATFORMS.
image 			    Build docker images for host arch.
image.multiarch     Build docker images for multiple platforms. See option PLATFORMS.
push                Build docker images for host arch and push images to registry.
push.multiarch      Build docker images for multiple platforms and push images to registry.
deploy 				Deploy updated components to development env.
clean 				Remove all files that are created by building.
lint 				Check syntax and styling of go sources.
test 				Run unit test.
cover 				Run unit test and get test coverage.
release 			Release iam.
format 				Gofmt (reformat) package sources (exclude vendor dir if existed).
verify-copyright 	Verify the boilerplate headers for all files.
add-copyright 		Ensures source code files have copyright license headers.
gen 				Generate all necessary files, such as error code files.
ca 					Generate CA files for all iam components.
install 			Install iam system with all its components.
swagger 			Generate swagger document.
serve-swagger 		Serve swagger spec and docs.
dependencies 		Install necessary dependencies.
tools 				install dependent tools.
check-updates 		Check outdated dependencies of the go projects.
help 				Show this help info.
```

#### 自动生成代码 

低代码的理念现在越来越流行。虽然低代码有很多缺点，但确实有很多优点，例如：

- 自动化生成代码，减少工作量，提高工作效率。 
- 代码有既定的生成规则，相比人工编写代码，准确性更高、更规范。

目前来看，自动生成代码现在已经成为趋势，比如 Kubernetes 项目有很多代码都是自动生成的。想写出一个优雅的 Go 项目，也应该认真思考哪些地方的代码可以自动生成。

在这门课的 IAM 项目中，就有大量的代码是自动生成的，放在这里供参考：

- 错误码、错误码说明文档。 
- 自动生成缺失的 doc.go 文件。 
- 利用 gotests 工具，自动生成单元测试用例。 
- 使用 Swagger 工具，自动生成 Swagger 文档。 
- 使用 Mock 工具，自动生成接口的 Mock 实例。

#### 善于借助工具 

在开发 Go 项目的过程中，也要善于借助工具，来完成一部分工作。利用工具可以带来很多好处：

- 解放双手，提高工作效率。 
- 利用工具的确定性，可以确保执行结果的一致性。例如，使用 golangci-lint 对代码进行检查，可以确保不同开发者开发的代码至少都遵循 golangci-lint 的代码检查规范。 
- 有利于实现自动化，可以将工具集成到 CI/CD 流程中，触发流水线自动执行。

那么，Go 项目中，有哪些工具可以为我们所用呢？这里，整理了一些有用的工具：

![image-20211111011449792](IAM-document.assets/image-20211111011449792.png)

所有这些工具都可以通过下面的方式安装。

```sh
$ cd $IAM_ROOT
$ make tools.install
```

IAM 项目使用了上面这些工具的绝大部分，用来尽可能提高整个项目的自动化程度，提高项目维护效率。 

#### 对接 CI/CD 

代码在合并入主干时，应该有一套 CI/CD 流程来自动化地对代码进行检查、编译、单元测试等，只有通过后的代码才可以并入主干。通过 CI/CD 流程来保证代码的质量。

当前比较流行的 CI/CD 工具有 Jenkins、GitLab、Argo、Github Actions、JenkinsX 等。后续会详细介绍 CI/CD 的原理和实战。 

### 编写高质量的项目文档 

最后，一个优雅的项目，还应该有完善的文档。例如 README.md、安装文档、开发文档、使用文档、API 接口文档、设计文档等等。这些内容在《规范设计之文档规范》有详细介绍。

### 总结 

使用 Go 语言做项目开发，核心目的其实就是开发一个优雅的 Go 项目。那么如何开发一个优雅的 Go 项目呢？

Go 项目包含三大内容，即 Go 应用、项目管理、项目文档，因此开发一个优雅的 Go 项目，其实就是编写高质量的 Go 应用、高效管理项目和编写高质量的项目文档。针对每一项，都给出了一些实现方式，这些方式详见下图：

![image-20211111012117058](IAM-document.assets/image-20211111012117058.png)

### 课后练习 

- 在工作中，还有哪些方法，来帮助开发一个优雅的 Go 项目呢？ 
- 在当前项目中有哪些可以接口化的代码呢？找到它们，并尝试用面向接口的编程哲学去重写这部分代码。



## Go 设计模式之GoF

Go 项目开发中常用的设计模式。 

在软件开发中，经常会遇到各种各样的编码场景，这些场景往往重复发生，因此具有典型性。针对这些典型场景，可以自己编码解决，也可以采取更为省时省力的方式：直接采用设计模式。 

设计模式是啥呢？简单来说，就是将软件开发中需要重复性解决的编码场景，按最佳实践的方式抽象成一个模型，模型描述的解决方法就是设计模式。使用设计模式，可以使代码更易于理解，保证代码的重用性和可靠性。 

在软件领域，GoF（四人帮，全拼 Gang of Four）首次系统化提出了 3 大类、共 25 种可复用的经典设计方案，来解决常见的软件设计问题，为可复用软件设计奠定了一定的理论基础。 

从总体上说，这些设计模式可以分为创建型模式、结构型模式、行为型模式 3 大类，用来完成不同的场景。

这一讲，会介绍几个在 Go 项目开发中比较常用的设计模式，用更加简单快捷的方法应对不同的编码场景。其中，简单工厂模式、抽象工厂模式和工厂 方法模式都属于工厂模式，会把它们放在一起讲解。

![image-20211111220908247](IAM-document.assets/image-20211111220908247.png)

### 创建型模式 

首先来看创建型模式（Creational Patterns），它提供了一种在创建对象的同时隐藏创建逻辑的方式，而不是使用 new 运算符直接实例化对象。 

这种类型的设计模式里，单例模式和工厂模式（具体包括简单工厂模式、抽象工厂模式和 工厂方法模式三种）在 Go 项目开发中比较常用。先来看单例模式。 

#### 单例模式

单例模式（Singleton Pattern），是最简单的一个模式。在 Go 中，单例模式指的是全局只有一个实例，并且它负责创建自己的对象。单例模式不仅有利于减少内存开支，还有减少系统性能开销、防止多个实例产生冲突等优点。 

因为单例模式保证了实例的全局唯一性，而且只被初始化一次，所以比较适合全局共享一个实例，且只需要被初始化一次的场景，例如数据库实例、全局配置、全局任务池等。

单例模式又分为饿汉方式和懒汉方式。

- 饿汉方式指全局的单例实例在包被加载时创建，
- 懒汉方式指全局的单例实例在第一次被使用时创建。

可以看到，这种命名方式非常形象地体现了它们不同的特点。

##### 饿汉方式

 接下来，就来分别介绍下这两种方式。先来看饿汉方式。 

下面是一个饿汉方式的单例模式代码：

```go
package hungry_singleton

type singleton struct {
}

var ins *singleton = &singleton{}

func GetInsOr() *singleton {
	return ins
}
```

需要注意，因为实例是在包被导入时初始化的，所以如果初始化耗时，会导致程序加载时间比较长。 

##### 懒汉方式

懒汉方式是开源项目中使用最多的，但它的缺点是非并发安全，在实际使用时需要加锁。 

以下是懒汉方式不加锁的一个实现：

```go
package hungry_singleton_unlock

type singleton struct {
}

var ins *singleton

func GetInsOr() *singleton {
	if ins == nil {
		ins = &singleton{}
	}
	return ins
}
```

可以看到，在创建 ins 时，如果 ins==nil，就会再创建一个 ins 实例，这时候单例就会有多个实例。 

为了解决懒汉方式非并发安全的问题，需要对实例进行加锁，下面是带检查锁的一个实现：

```go
package hungry_singleton_lock

import "sync"

type singleton struct {
}

var ins *singleton
var mu sync.Mutex

func GetIns() *singleton {
   if ins == nil {
      mu.Lock()
      if ins == nil {
         ins = &singleton{}
      }
      mu.Unlock()
   }
   return ins
}
```

上述代码只有在创建时才会加锁，既提高了代码效率，又保证了并发安全。 

##### once.Do 方式

除了饿汉方式和懒汉方式，在 Go 开发中，还有一种更优雅的实现方式，建议采用这种方式，代码如下：

```go
package onceDo

import (
   "sync"
)

type singleton struct {
}

var ins *singleton
var once sync.Once

func GetInsOr() *singleton {
   once.Do(func() {
      ins = &singleton{}
   })
   return ins
}
```

使用once.Do可以确保 ins 实例全局只被创建一次，once.Do 函数还可以确保当同时有多个创建动作时，只有一个创建动作在被执行。 

另外，IAM 应用中大量使用了单例模式，如果想了解更多单例模式的使用方式，可以直接查看 IAM 项目代码。IAM 中单例模式有 GetStoreInsOr、GetEtcdFactoryOr、 GetMySQLFactoryOr、GetCacheInsOr等。

#### 工厂模式 

工厂模式（Factory Pattern）是面向对象编程中的常用模式。在 Go 项目开发中，可以通过使用多种不同的工厂模式，来使代码更简洁明了。

Go 中的结构体，可以理解为面向对象编程中的类，例如 Person 结构体（类）实现了 Greet 方法。

```go
package FactoryPattern

import "fmt"

type Person struct {
	Name string
	Age  int
}

func (p Person) Greet() {
	fmt.Printf("Hi! My name is %s", p.Name)
}
```

有了 Person“类”，就可以创建 Person 实例。可以通过简单工厂模式、抽象工厂模 式、工厂方法模式这三种方式，来创建一个 Person 实例。 

##### 简单工厂模式

这三种工厂模式中，简单工厂模式是最常用、最简单的。它就是一个接受一些参数，然后返回 Person 实例的函数：

```go
package SimpleFactoryPattern

import "fmt"

type Person struct {
   Name string
   Age  int
}

func (p Person) Greet() {
   fmt.Printf("Hi! My name is %s", p.Name)
}
func NewPerson(name string, age int) *Person {
   return &Person{
      Name: name,
      Age:  age,
   }
}
```

和p：=＆Person {}这种创建实例的方式相比，简单工厂模式可以确保创建的实例具有需要的参数，进而保证实例的方法可以按预期执行。例如，通过NewPerson创建 Person 实例时，可以确保实例的 name 和 age 属性被设置。 

##### 抽象工厂模式

再来看抽象工厂模式，它和简单工厂模式的唯一区别，就是它返回的是接口而不是结构体。 

通过返回接口，可以在不公开内部实现的情况下，让调用者使用提供的各种功能，例如：

```go
package AbstractFactoryPattern

import "fmt"

type Person interface {
   Greet()
}
type person struct {
   name string
   age  int
}

func (p person) Greet() {
   fmt.Printf("Hi! My name is %s", p.name)
}

// NewPerson Here, NewPerson returns an interface, and not the person struct itself
func NewPerson(name string, age int) Person {
   return person{
      name: name,
      age:  age,
   }
}
```

上面这个代码，定义了一个不可导出的结构体person，在通过 NewPerson 创建实例的时候返回的是接口，而不是结构体。 

通过返回接口，还可以实现多个工厂函数，来返回不同的接口实现，例如：

```go
package HTTPClientFactoryPattern

import (
   "net/http"
   "net/http/httptest"
)

// We define a Doer interface, that has the method signature
// of the `http.Client` structs `Do` method
type Doer interface {
   Do(req *http.Request) (*http.Response, error)
}

// This gives us a regular HTTP client from the `net/http` package
func NewHTTPClient() Doer {
   return &http.Client{}
}

type mockHTTPClient struct{}

func (*mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
   // The `NewRecorder` method of the httptest package gives us
   // a new mock request generator
   res := httptest.NewRecorder()
   // calling the `Result` method gives us
   // the default empty *http.Response object
   return res.Result(), nil
}

// This gives us a mock HTTP client, which returns
// an empty response for any request sent to it
func NewMockHTTPClient() Doer {
   return &mockHTTPClient{}
}
```

NewHTTPClient和NewMockHTTPClient都返回了同一个接口类型 Doer，这使得二者可以互换使用。

当想测试一段调用了 Doer 接口 Do 方法的代码时，这一点特别有用。可以使用一个 Mock 的 HTTP 客户端，从而避免了调用真实外部接口可能带来的失败。

来看个例子，假设想测试下面这段代码：

```go
package HTTPClientFactoryPattern

import "net/http"

// QueryUser testing
func QueryUser(doer Doer) error {
   req, err := http.NewRequest("Get", "https://iam.api.marmotedu.com:8080/v1/secrets", nil)
   if err != nil {
      return err
   }
   _, err = doer.Do(req)
   if err != nil {
      return err
   }
   return nil
}
```

其测试用例为：

```go
func TestQueryUser(t *testing.T) {
   doer := NewMockHTTPClient()
   if err := QueryUser(doer); err != nil {
      t.Errorf("QueryUser failed, err: %v", err)
   }
}
```

另外，在使用简单工厂模式和抽象工厂模式返回实例对象时，都可以返回指针。例如，简单工厂模式可以这样返回实例对象：

```go
return &Person{
  Name: name,
  Age:  age,
}
```

抽象工厂模式可以这样返回实例对象：

```go
return &person{
   name: name,
   age:  age,
}
```

在实际开发中，建议返回非指针的实例，因为主要是想通过创建实例，调用其提供的方法，而不是对实例做更改。如果需要对实例做更改，可以实现 `SetXXX` 的方法。通过返回非指针的实例，可以确保实例的属性，避免属性被意外 / 任意修改。 

在简单工厂模式中，依赖于唯一的工厂对象，如果需要实例化一个产品，就要向工厂中传入一个参数，获取对应的对象；如果要增加一种产品，就要在工厂中修改创建产品的函数。这会导致耦合性过高，这时就可以使用工厂方法模式。 

##### 工厂方法模式

在工厂方法模式中，依赖工厂接口，可以通过实现工厂接口来创建多种工厂，将对象创建从由一个对象负责所有具体类的实例化，变成由一群子类来负责对具体类的实例化， 从而将过程解耦。 

下面是工厂方法模式的一个代码实现：

```go
package FactoryMethodFactoryPattern

type Person struct {
   name string
   age  int
}

func NewPersonFactory(age int)func(name string) Person {
   return func(name string) Person {
      return Person{
         name: name,
         age:  age,
      }
   }
}
```

然后，可以使用此功能来创建具有默认年龄的工厂：

```go
func main() {
   newBaby := NewPersonFactory(1)
   baby := newBaby("john")
   fmt.Println("baby is", baby.name, "age is", baby.age)

   newTeenager := NewPersonFactory(16)
   teen := newTeenager("jill")
   fmt.Println("teenager is", teen.name, "age is", teen.age)
}
```



### 结构型模式 

已经介绍了单例模式、工厂模式这两种创建型模式，接下来来看结构型模式 （Structural Patterns），它的特点是关注类和对象的组合。这一类型里，详细讲讲策略模式和模板模式。

#### 策略模式

策略模式（Strategy Pattern）定义一组算法，将每个算法都封装起来，并且使它们之间可以互换。 

在什么时候，需要用到策略模式呢？ 

在项目开发中，经常要根据不同的场景，采取不同的措施，也就是不同的策略。比如，假设需要对 a、b 这两个整数进行计算，根据条件的不同，需要执行不同的计算方式。可以把所有的操作都封装在同一个函数中，然后通过 if ... else ... 的形式来调用不同的计算方式，这种方式称之为硬编码。 

在实际应用中，随着功能和体验的不断增长，需要经常添加 / 修改策略，这样就需要不断修改已有代码，不仅会让这个函数越来越难维护，还可能因为修改带来一些 bug。所以为了解耦，需要使用策略模式，定义一些独立的类来封装不同的算法，每一个类封装一个具体的算法（即策略）。 

下面是一个实现策略模式的代码：

```go
package StrategyPattern

// IStrategy 策略模式
// 定义一个策略类
type IStrategy interface {
   do(int, int) int
}

// 策略实现：加
type add struct{}

func (*add) do(a, b int) int {
   return a + b
}

// 策略实现：减
type reduce struct{}

func (*reduce) do(a, b int) int {
   return a - b
}

// Operator 具体策略的执行者
type Operator struct {
   strategy IStrategy
}

// 设置策略
func (operator *Operator) setStrategy(strategy IStrategy) {
   operator.strategy = strategy
}

// 调用策略中的方法
func (operator *Operator) calculate(a, b int) int {
   return operator.strategy.do(a, b)
}
```

在上述代码中，定义了策略接口 IStrategy，还定义了 add 和 reduce 两种策略。最后定义了一个策略执行者，可以设置不同的策略，并执行，例如：

```go
package StrategyPattern

import (
   "fmt"
   "testing"
)

// TestStrategy 执行测试
func TestStrategy(t *testing.T) {
   operator := Operator{}
   operator.setStrategy(&add{})
   result := operator.calculate(1, 2)
   fmt.Println("add:", result)

   operator.setStrategy(&reduce{})
   result = operator.calculate(2, 1)
   fmt.Println("reduce:", result)
}
```

可以看到，可以随意更换策略，而不影响 Operator 的所有实现。

#### 模版模式 

模版模式 (Template Pattern) 定义一个操作中算法的骨架，而将一些步骤延迟到子类中。 这种方法让子类在不改变一个算法结构的情况下，就能重新定义该算法的某些特定步骤。 

简单来说，模板模式就是将一个类中能够公共使用的方法放置在抽象类中实现，将不能公共使用的方法作为抽象方法，强制子类去实现，这样就做到了将一个类作为一个模板，让开发者去填充需要填充的地方。 

以下是模板模式的一个实现：

```go
package TemplatePattern

import "fmt"

type Cooker interface {
   fire()
   cooke()
   outfire()
}

// CookMenu 类似于一个抽象类
type CookMenu struct {
}

func (CookMenu) fire() {
   fmt.Println("开火")
}

// 做菜，交给具体的子类实现
func (CookMenu) cooke() {
}

func (CookMenu) outfire() {
   fmt.Println("关火")
}

// 封装具体步骤
func doCook(cook Cooker) {
   cook.fire()
   cook.cooke()
   cook.outfire()
}

type XiHongShi struct {
   CookMenu
}

func (*XiHongShi) cooke() {
   fmt.Println("做西红柿")
}

type ChaoJiDan struct {
   CookMenu
}

func (ChaoJiDan) cooke() {
   fmt.Println("做炒鸡蛋")
}
```

这里来看下测试用例：

```go
package TemplatePattern

import (
   "fmt"
   "testing"
)

func TestTemplate(t *testing.T) {
   // 做西红柿
   xihongshi := &XiHongShi{}
   doCook(xihongshi)
   fmt.Println("\n=====> 做另外一道菜")

   // 做炒鸡蛋
   chaojidan := &ChaoJiDan{}
   doCook(chaojidan)
}
```



### 行为型模式

行为型模式（Behavioral Patterns），它的特点是关注对象之间的通信。这一类别的设计模式中，会讲到代理模式和选项模式。 

#### 代理模式 

代理模式 (Proxy Pattern)，可以为另一个对象提供一个替身或者占位符，以控制对这个对象的访问。

以下代码是一个代理模式的实现：

```go
package main

import "fmt"

type Seller interface {
   sell(name string)
}

// Station 火车站
type Station struct {
   stock int //库存
}

func (station *Station) sell(name string) {
   if station.stock > 0 {
      station.stock--
      fmt.Printf("代理点中：%s买了一张票,剩余：%d \n", name, station.stock)
   } else {
      fmt.Println("票已售空")
   }
}

// StationProxy 火车代理点
type StationProxy struct {
   station *Station // 持有一个火车站对象
}

func (proxy *StationProxy) sell(name string) {
   if proxy.station.stock > 0 {
      proxy.station.stock--
      fmt.Printf("代理点中：%s买了一张票,剩余：%d \n", name, proxy.station.stock)
   } else {
      fmt.Println("票已售空")
   }
}
```

上述代码中，StationProxy 代理了 Station，代理类中持有被代理类对象，并且和被代理类对象实现了同一接口。 

#### 选项模式 

选项模式（Options Pattern）也是 Go 项目开发中经常使用到的模式，例如，grpc/grpcgo 的 NewServer 函数，uber-go/zap 包的 New 函数都用到了选项模式。

使用选项模式，可以创建一个带有默认值的 struct 变量，并选择性地修改其中一些参数的值。

在 Python 语言中，创建一个对象时，可以给参数设置默认值，这样在不传入任何参数时，可以返回携带默认值的对象，并在需要时修改对象的属性。这种特性可以大大简化开发者创建一个对象的成本，尤其是在对象拥有众多属性时。 

而在 Go 语言中，因为不支持给参数设置默认值，为了既能够创建带默认值的实例，又能够创建自定义参数的实例，不少开发者会通过以下两种方法来实现： 

第一种方法，要分别开发两个用来创建实例的函数，一个可以创建带默认值的实例， 一个可以定制化创建实例。

```go
package main

import (
   "time"
)

const (
   defaultCaching = false
   defaultTimeout = 10
)

type Connection struct {
   addr    string
   cache   bool
   timeout time.Duration
}

// NewConnect creates a connection.
func NewConnect(addr string) (*Connection, error) {
   return &Connection{
      addr:    addr,
      cache:   defaultCaching,
      timeout: defaultTimeout,
   }, nil
}

// NewConnectWithOptions creates a connection with options.
func NewConnectWithOptions(addr string, cache bool, timeout time.Duration) (*Connection, error) {
   return &Connection{
      addr:    addr,
      cache:   cache,
      timeout: timeout,
   }, nil
}
```

使用这种方式，创建同一个 Connection 实例，却要实现两个不同的函数，实现方式很不 优雅。 

另外一种方法相对优雅些。需要创建一个带默认值的选项，并用该选项创建实例：

```go
package OneDefaultOptionsPattern

import (
   "time"
)

const (
   defaultTimeout = 10
   defaultCaching = false
)

type Connection struct {
   addr    string
   cache   bool
   timeout time.Duration
}
type ConnectionOptions struct {
   Caching bool
   Timeout time.Duration
}

func NewDefaultOptions() *ConnectionOptions {
   return &ConnectionOptions{
      Caching: defaultCaching,
      Timeout: defaultTimeout,
   }
}

// NewConnect creates a connection with options.
func NewConnect(addr string, opts *ConnectionOptions) (*Connection, error) {
   return &Connection{
      addr:    addr,
      cache:   opts.Caching,
      timeout: opts.Timeout,
   }, nil
}
```

使用这种方式，虽然只需要实现一个函数来创建实例，但是也有缺点：为了创建 Connection 实例，每次都要创建 ConnectionOptions，操作起来比较麻烦。

那么有没有更优雅的解决方法呢？答案当然是有的，就是使用选项模式来创建实例。以下代码通过选项模式实现上述功能：

```go
package OptionsPattern

import (
   "time"
)

type Connection struct {
   addr    string
   cache   bool
   timeout time.Duration
}

const (
   defaultTimeout = 10
   defaultCaching = false
)

type options struct {
   timeout time.Duration
   caching bool
}

// Option overrides behavior of Connect.
type Option interface {
   apply(*options)
}
type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
   f(o)
}
func WithTimeout(t time.Duration) Option {
   return optionFunc(func(o *options) {
      o.timeout = t
   })
}
func WithCaching(cache bool) Option {
   return optionFunc(func(o *options) {
      o.caching = cache
   })
}

// NewConnect Connect creates a connection.
func NewConnect(addr string, opts ...Option) (*Connection, error) {
   options := options{
      caching: defaultCaching,
      timeout: defaultTimeout,
   }
   for _, o := range opts {
      o.apply(&options)
   }
   return &Connection{
      addr:    addr,
      cache:   options.caching,
      timeout: options.timeout,
   }, nil
}
```

在上面的代码中，首先定义了options结构体，它携带了 timeout、caching 两个属性。接下来，通过NewConnect创建了一个连接，NewConnect 函数中先创建了一个带有默认值的options结构体变量，并通过调用

```go
for _, o := range opts {
   o.apply(&options)
}
```

来修改所创建的options结构体变量。 

需要修改的属性，是在NewConnect时，通过 Option 类型的选项参数传递进来的。可以通过WithXXX函数来创建 Option 类型的选项参数：WithTimeout、WithCaching。

Option 类型的选项参数需要实现 apply(*options)函数，结合 WithTimeout、 WithCaching 函数的返回值和 optionFunc 的 apply 方法实现，可以知道 o.apply(&options)其实就是把 WithTimeout、WithCaching 传入的参数赋值给 options 结构体变量，以此动态地设置 options 结构体变量的属性。 

这里还有一个好处：可以在 apply 函数中自定义赋值逻辑，例如o.timeout = 100 * t。通过这种方式，会有更大的灵活性来设置结构体的属性。

选项模式有很多优点，例如：

- 支持传递多个参数，并且在参数发生变化时保持兼容性；
- 支持任意顺序传递参数；
- 支持默认值；
- 方便扩展；
- 通过 WithXXX 的函数命名，可以使参数意义更加明确，等等。 

不过，为了实现选项模式，增加了很多代码，所以在开发中，要根据实际场景选择是否使用选项模式。选项模式通常适用于以下场景：

- 结构体参数很多，创建结构体时，期望创建一个携带默认值的结构体变量，并选择性修改其中一些参数的值。 
- 结构体参数经常变动，变动时又不想修改创建实例的函数。例如：结构体新增一个 retry 参数，但是又不想在 NewConnect 入参列表中添加retry int这样的参数声明。

如果结构体参数比较少，可以慎重考虑要不要采用选项模式。 

### 总结 

设计模式，是业界沉淀下来的针对特定场景的最佳解决方案。在软件领域，GoF 首次系统化提出了 3 大类设计模式：创建型模式、结构型模式、行为型模式。 

这一讲，介绍了 Go 项目开发中 6 种常用的设计模式。每种设计模式解决某一类场景， 总结成了一张表格。

![image-20211112230630944](IAM-document.assets/image-20211112230630944.png)

### 课后练习

- 当前开发的项目中，哪些可以用单例模式、工厂模式、选项模式来重新实现呢？如果有的话，试着重写下这部分代码。
- 除了这 6 种设计模式之外，还用过其他的设计模式吗？



## Go 编码规范

一份清晰、可直接套用的 Go 编码规范，编写一个高质量的 Go 应用。 

这份规范，参考了 Go 官方提供的编码规范，以及 Go 社区沉淀的一些比较合理的规范，加入自己的理解总结出的，它比很多公司内部的规范更全面。

这份编码规范中包含代码风格、命名规范、注释规范、类型、控制结构、函数、GOPATH 设置规范、依赖管理和最佳实践九类规范，作为写代码时候的一个参考手册。 

### 代码风格 

#### 代码格式

代码都必须用 go fmt 进行格式化。 

运算符和操作数之间要留空格。 

建议一行代码不超过 120 个字符，超过部分，请采用合适的换行方式换行。但也有些例外场景，例如 import 行、工具自动生成的代码、带 tag 的 struct 字段。 

文件长度不能超过 800 行。 

函数长度不能超过 80 行。 

import 规范 

- 代码都必须用 go imports 进行格式化（建议将代码 Go 代码编辑器设置为：保存时运行 go imports）。 

- 不要使用相对路径引入包，例如 import ../util/net 。 

- 包名称与导入路径的最后一个目录名不匹配时，或者多个相同包名冲突时，则必须使用导入别名。

  - ```go
    // bad
    "github.com/dgrijalva/jwt-go/v4"
    
    //good
    jwt "github.com/dgrijalva/jwt-go/v4"
    ```

- 导入的包建议进行分组，匿名包的引用使用一个新的分组，并对匿名包引用进行说 明。

  - ```go
    import (
    	// go 标准包
    	"fmt"
    	
    	// 第三方包
    	"github.com/jinzhu/gorm"
    	"github.com/spf13/cobra"
    	"github.com/spf13/viper"
    	
    	// 匿名包单独分组，并对匿名包引用进行说明
    	// import mysql driver
    	_ "github.com/jinzhu/gorm/dialects/mysql"
    	
    	// 内部包
    	v1 "github.com/marmotedu/api/apiserver/v1"
    	metav1 "github.com/marmotedu/apimachinery/pkg/meta/v1"
    	"github.com/marmotedu/iam/pkg/cli/genericclioptions"
    )
    ```

####  声明、初始化和定义

当函数中需要使用到多个变量时，可以在函数开始处使用 var 声明。在函数外部声明必须使用 var ，不要采用 := ，容易踩到变量的作用域的问题。

```go
var (
  Width int
  Height int
)
```

在初始化结构引用时，请使用 &T{}代替 new(T)，以使其与结构体初始化一致。

```go
// bad
sptr := new(T)
sptr.Name = "bar"

// good
sptr := &T{Name: "bar"}
```

struct 声明和初始化格式采用多行，定义如下。

```go
type User struct{
  Username string
  Email string
}

user := User{
  Username: "colin",
  Email: "colin404@foxmail.com",
}
```

相似的声明放在一组，同样适用于常量、变量和类型声明。

```go
// bad
import "a"
import "b"

// good
import (
  "a"
  "b"
)
```

尽可能指定容器容量，以便为容器预先分配内存，例如：

```go
v := make(map[int]string, 4)
v := make([]string, 0, 4)
```

在顶层，使用标准 var 关键字。请勿指定类型，除非它与表达式的类型不同。

```go
// bad
var _s string = F()
func F() string { return "A" }

// good
var _s = F()
// 由于 F 已经明确了返回一个字符串类型，因此我们没有必要显式指定_s 的类型
// 还是那种类型
func F() string { return "A" }
```

对于未导出的顶层常量和变量，使用 _ 作为前缀。

```go
// bad
const (
  defaultHost = "127.0.0.1"
  defaultPort = 8080
)

// good
const (
  _defaultHost = "127.0.0.1"
  _defaultPort = 8080
)
```

嵌入式类型（例如 mutex）应位于结构体内的字段列表的顶部，并且必须有一个空行将嵌入式字段与常规字段分隔开。

```go
// bad
type Client struct {
  version int
  http.Client
}

// good
type Client struct {
  http.Client
  version int
}
```

#### 错误处理

- error 作为函数的值返回，必须对error进行处理，或将返回值赋值给明确忽略。对于 defer xx.Close()可以不用显式处理。

```go
func load() error {
	// normal code
}

// bad
load()

// good
_ = load()
```

- error 作为函数的值返回且有多个返回值的时候，error必须是最后一个参数。

```go
// bad
func load() (error, int) {
	// normal code
}

// good
func load() (int, error) {
	// normal code
}
```

- 尽早进行错误处理，并尽早返回，减少嵌套。

```go
// bad
if err != nil {
  	// error code
  } else {
  	// normal code
}

// good
if err != nil {
  // error handling
  return err
}
// normal code
```

- 如果需要在 if 之外使用函数调用的结果，则应采用下面的方式。

```go
// bad
if v, err := foo(); err != nil {
	// error handling
}

// good
v, err := foo()
if err != nil {
	// error handling
}
```

- 错误要单独判断，不与其他逻辑组合判断。

```go
// bad
v, err := foo()
if err != nil || v == nil {
  // error handling
  return err
}

// good
v, err := foo()
if err != nil {
  // error handling
  return err
}
if v == nil {
  // error handling
  return errors.New("invalid value v")
}
```

- 如果返回值需要初始化，则采用下面的方式。

```go
v, err := f()
if err != nil {
  // error handling
  return // or continue.
}
// use v
```

- 错误描述建议

  - 告诉用户他们可以做什么，而不是告诉他们不能做什么。 

  - 当声明一个需求时，用 must 而不是 should。例如，must be greater than 0、 must match regex '[a-z]+'。 

  - 当声明一个格式不对时，用 must not。例如，must not contain。 

  - 当声明一个动作时用 may not。例如，may not be specified when otherField is empty、only name may be specified。 

  - 引用文字字符串值时，请在单引号中指示文字。例如，ust not contain '..'。 

  - 当引用另一个字段名称时，请在反引号中指定该名称。例如，must be greater than request。 

  - 指定不等时，请使用单词而不是符号。例如，must be less than 256、must be greater than or equal to 0 (不要用 larger than、bigger than、more than、 higher than)。 

  - 指定数字范围时，请尽可能使用包含范围。 建议 Go 1.13 以上，error 生成方式为 fmt.Errorf("module xxx: %w", err)。 

  - 错误描述用小写字母开头，结尾不要加标点符号，例如：

    - ```go
      // bad
      errors.New("Redis connection failed")
      errors.New("redis connection failed.")
      
      // good
      errors.New("redis connection failed")
      ```

####  panic 处理

在业务逻辑处理中禁止使用 panic。 

在 main 包中，只有当程序完全不可运行时使用 panic，例如无法打开文件、无法连接数据库导致程序无法正常运行。

在 main 包中，使用 log.Fatal 来记录错误，这样就可以由 log 来结束程序，或者将 panic 抛出的异常记录到日志文件中，方便排查问题。 

可导出的接口一定不能有 panic。 

包内建议采用 error 而不是 panic 来传递错误。

#### 单元测试

单元测试文件名命名规范为 example_test.go。 

每个重要的可导出函数都要编写测试用例。 因为单元测试文件内的函数都是不对外的，所以可导出的结构体、函数等可以不带注释。 

如果存在 func (b *Bar) Foo ，单测函数可以为 func TestBar_Foo。

#### 类型断言失败处理

type assertion 的单个返回值针对不正确的类型将产生 panic。请始终使用 “comma ok”的惯用法。

```go
// bad
t := n.(int)

// good
t, ok := n.(int)
if !ok {
	// error handling
}
// normal code
```

### 命名规范

命名规范是代码规范中非常重要的一部分，一个统一的、短小的、精确的命名规范可以大大提高代码的可读性，也可以借此规避一些不必要的 Bug。

#### 包命名

包名必须和目录名一致，尽量采取有意义、简短的包名，不要和标准库冲突。

包名全部小写，没有大写或下划线，使用多级目录来划分层级。 

项目名可以通过中划线来连接多个单词。 

包名以及包所在的目录名，不要使用复数，例如，是net/utl，而不是net/urls。 

不要用 common、util、shared 或者 lib 这类宽泛的、无意义的包名。 

包名要简单明了，例如 net、time、log。

#### 函数命名

函数名采用驼峰式，首字母根据访问控制决定使用大写或小写，例如：MixedCaps 或者 mixedCaps。 

代码生成工具自动生成的代码 (如 xxxx.pb.go) 和为了对相关测试用例进行分组，而采用的下划线 (如 TestMyFunction_WhatIsBeingTested) 排除此规则。

#### 文件命名

文件名要简短有意义。 

文件名应小写，并使用下划线分割单词。

#### 结构体命名

采用驼峰命名方式，首字母根据访问控制决定使用大写或小写，例如 MixedCaps 或者 mixedCaps。 

结构体名不应该是动词，应该是名词，比如 Node、NodeSpec。 

避免使用 Data、Info 这类无意义的结构体名。 

结构体的声明和初始化应采用多行，例如：

```go
// User 多行声明
type User struct {
  Name string
  Email string
}

// 多行初始化
u := User{
	UserName: "colin",
  Email: "colin404@foxmail.com",
}
```

#### 接口命名

接口命名的规则，基本和结构体命名规则保持一致： 

- 单个函数的接口名以 “er"”作为后缀（例如 Reader，Writer），有时候可能导致蹩脚的英文，但是没关系。 
- 两个函数的接口名以两个函数名命名，例如 ReadWriter。 
- 三个以上函数的接口名，类似于结构体名。

例如：

```go
// Seeking to an offset before the start of the file is an error.
// Seeking to any positive offset is legal, but the behavior of subsequent
// I/O operations on the underlying object is implementation-dependent.
type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}

// ReadWriter is the interface that groups the basic Read and Write methods.
type ReadWriter interface {
  Reader
  Writer
}
```

#### 变量命名

变量名必须遵循驼峰式，首字母根据访问控制决定使用大写或小写。 

在相对简单（对象数量少、针对性强）的环境中，可以将一些名称由完整单词简写为单个字母，例如： 

- user 可以简写为 u； 
- userID 可以简写 uid。 

特有名词时，需要遵循以下规则： 

- 如果变量为私有，且特有名词为首个单词，则使用小写，如 apiClient。
- 其他情况都应当使用该名词原有的写法，如 APIClient、repoID、UserID。

下面列举了一些常见的特有名词。

```go
// A GonicMapper that contains a list of common initialisms taken from golang/lint
var LintGonicMapper = GonicMapper{
  "API": true,
  "ASCII": true,
  "CPU": true,
  "CSS": true,
  "DNS": true,
  "EOF": true,
  "GUID": true,
  "HTML": true,
  "HTTP": true,
  "HTTPS": true,
  "ID": true,
  "IP": true,
  "JSON": true,
  "LHS": true,
  "QPS": true,
  "RAM": true,
  "RHS": true,
  "RPC": true,
  "SLA": true,
  "SMTP": true,
  "SSH": true,
  "TLS": true,
  "TTL": true,
  "UI": true,
  "UID": true,
  "UUID": true,
  "URI": true,
  "URL": true,
  "UTF8": true,
  "VM": true,
  "XML": true,
  "XSRF": true,
  "XSS": true,
}
```

若变量类型为 bool 类型，则名称应以 Has，Is，Can 或 Allow 开头，例如：

```go
var hasConflict bool
var isExist bool
var canManage bool
var allowGitHook bool
```

局部变量应当尽可能短小，比如使用 buf 指代 buffer，使用 i 指代 index。 

代码生成工具自动生成的代码可排除此规则 (如 xxx.pb.go 里面的 Id)

#### 常量命名

常量名必须遵循驼峰式，首字母根据访问控制决定使用大写或小写。 

如果是枚举类型的常量，需要先创建相应类型：

```go
// Code defines an error code type.
type Code int

// Internal errors.
const (
  // ErrUnknown - 0: An unknown error occurred.
  ErrUnknown Code = iota
  // ErrFatal - 1: An fatal error occurred.
  ErrFatal
)
```

#### Error 的命名

Error 类型应该写成 FooError 的形式。

```go
type ExitError struct {
	// ....
}
```

Error 变量写成 ErrFoo 的形式。

```go
var ErrFormat = errors.New("unknown format")
```

### 注释规范

每个可导出的名字都要有注释，该注释对导出的变量、函数、结构体、接口等进行简要介绍。 

全部使用单行注释，禁止使用多行注释。 

和代码的规范一样，单行注释不要过长，禁止超过 120 字符，超过的请使用换行展示， 尽量保持格式优雅。 

注释必须是完整的句子，以需要注释的内容作为开头，句点作为结尾，格式为 // 名称 描述. 。例如：

```go
// bad
// logs the flags in the flagset.
func PrintFlags(flags *pflag.FlagSet) {
  // normal code
}

// good
// PrintFlags logs the flags in the flagset.
func PrintFlags(flags *pflag.FlagSet) {
	// normal code
}
```

所有注释掉的代码在提交 code review 前都应该被删除，否则应该说明为什么不删除， 并给出后续处理建议。 

在多段注释之间可以使用空行分隔加以区分，如下所示：

```go
// Package superman implements methods for saving the world.
//
// Experience has shown that a small number of procedures can prove
// helpful when attempting to save the world.
package superman
```

#### 包注释

每个包都有且仅有一个包级别的注释。

包注释统一用 // 进行注释，格式为 // Package 包名 包描述 ，例如：

```go
// Package genericclioptions contains flags which can be added to you command,
// useful helper functions.
package genericclioptions
```

#### 变量/常量注释

每个可导出的变量 / 常量都必须有注释说明，格式为// 变量名 变量描述，例如：

```go
// ErrSigningMethod defines invalid signing method error.
var ErrSigningMethod = errors.New("Invalid signing method")
```

出现大块常量或变量定义时，可在前面注释一个总的说明，然后在每一行常量的前一行或末尾详细注释该常量的定义，例如：

```go
// Code must start with 1xxxxx.
const (
  // ErrSuccess - 200: OK.
  ErrSuccess int = iota + 100001
  // ErrUnknown - 500: Internal server error.
  ErrUnknown
  // ErrBind - 400: Error occurred while binding the request body to the str
  ErrBind
  // ErrValidation - 400: Validation failed.
  ErrValidation
）
```

#### 结构体注释

每个需要导出的结构体或者接口都必须有注释说明，格式为 // 结构体名 结构体描述.。

结构体内的可导出成员变量名，如果意义不明确，必须要给出注释，放在成员变量的前一行或同一行的末尾。例如：

```go
// User represents a user restful resource. It is also used as gorm model.
type User struct {
  // Standard object's metadata.
  metav1.ObjectMeta `json:"metadata,omitempty"`
  Nickname string `json:"nickname" gorm:"column:nickname"`
  Password string `json:"password" gorm:"column:password"`
  Email string `json:"email" gorm:"column:email"`
  Phone string `json:"phone" gorm:"column:phone"`
  IsAdmin int `json:"isAdmin,omitempty" gorm:"column:isAdmin"`
}
```

#### 方法注释

每个需要导出的函数或者方法都必须有注释，格式为// 函数名 函数描述.，例如：

```go
// BeforeUpdate run before update database record.
func (p *Policy) BeforeUpdate() (err error) {
  // normal code
  return nil
}
```

#### 类型注释

每个需要导出的类型定义和类型别名都必须有注释说明，格式为 // 类型名 类型描述. ，例如：

```go
// Code defines an error code type.
type Code int
```

### 类型

#### 字符串

空字符串判断。

```go
// bad
if s == "" {
	// normal code
}

// good
if len(s) == 0 {
	// normal code
}
```

[]byte/string 相等比较。

```go
// bad
var s1 []byte
var s2 []byte
...
bytes.Equal(s1, s2) == 0
bytes.Equal(s1, s2) != 0

// good
var s1 []byte
var s2 []byte
...
bytes.Compare(s1, s2) == 0
bytes.Compare(s1, s2) != 0
```

字符串是否包含子串或字符。

```go
// bad
strings.Contains(s, subStr)
strings.ContainsAny(s, char)
strings.ContainRune(s, r)

// good
strings.Index(s, subStr) > -1
strings.IndexAny(s, char) > -1
strings.IndexRune(s, r) > -1
```

去除前后子串。

```go
// bad
var s1 = "a string value"
var s2 = "a "
var s3 = strings.TrimPrefix(s1, s2)

// good
var s1 = "a string value"
var s2 = "a "
var s3 string
if strings.HasPrefix(s1, s2) {
  s3 = s1[len(s2):]
}
```

复杂字符串使用 raw 字符串避免字符转义。

```go
// bad
regexp.MustCompile("\\.")

// good
regexp.MustCompile(`\.`)
```

#### 切片

空 slice 判断。

```go
// bad
if len(slice) = 0 {
  // normal code
}

// good
if slice != nil && len(slice) == 0 {
	// normal code
}
```

上面判断同样适用于 map、channel。

声明 slice。

```go
// bad
s := []string{}
s := make([]string, 0)

// good
var s []string
```

slice 复制。

```go
// bad
var b1, b2 []byte
for i, v := range b1 {
  b2[i] = v
}
for i := range b1 {
  b2[i] = b1[i]
}

// good
copy(b2, b1)
```

slice 新增。

```go
// bad
var a, b []int
for _, v := range a {
  b = append(b, v)
}

// good
var a, b []int
b = append(b, a...)
```

#### 结构体

struct 初始化。

struct 以多行格式初始化。

```go
type user struct {
  Id int64
  Name string
}

u1 := user{100, "Colin"}

u2 := user{
  Id: 200,
  Name: "Lex",
}
```

### 控制结构

#### if

if 接受初始化语句，约定如下方式建立局部变量。

```go
if err := loadConfig(); err != nil {
  // error handling
  return err
}
```

if 对于 bool 类型的变量，应直接进行真假判断。

```go
var isAllow bool
if isAllow {
	// normal code
}
```

#### for

采用短声明建立局部变量。

```go
sum := 0
for i := 0; i < 10; i++ {
  sum += 1
}
```

不要在 for 循环里面使用 defer，defer 只有在函数退出时才会执行。

```go
// bad
for file := range files {
  fd, err := os.Open(file)
  if err != nil {
    return err
  }
  defer fd.Close()
	// normal code
}

// good
for file := range files {
  func() {
    fd, err := os.Open(file)
    if err != nil {
      return err
    }
    defer fd.Close()
    // normal code
    }()
}
```

#### range

如果只需要第一项（key），就丢弃第二个。

```go
for key := range keys {
	// normal code
}
```

如果只需要第二项，则把第一项置为下划线。

```go
sum := 0
for _, value := range array {
	sum += value
}
```

#### switch 

必须要有 default。

```go
switch os := runtime.GOOS; os {
  case "linux":
  	fmt.Println("Linux.")
  case "darwin":
  	fmt.Println("OS X.")
  default:
  	fmt.Printf("%s.\n", os)
}
```

#### goto

业务代码禁止使用 goto 。 

框架或其他底层源码尽量不用。

### 函数

传入变量和返回变量以小写字母开头。 

函数参数个数不能超过 5 个。 

函数分组与顺序

- 函数应按粗略的调用顺序排序。 
- 同一文件中的函数应按接收者分组。 

尽量采用值传递，而非指针传递。 

传入参数是 map、slice、chan、interface ，不要传递指针。

#### 函数参数

如果函数返回相同类型的两个或三个参数，或者如果从上下文中不清楚结果的含义，使用命名返回，其他情况不建议使用命名返回，例如：

```go
func coordinate() (x, y float64, err error) {
	// normal code
}
```

传入变量和返回变量都以小写字母开头。 

尽量用值传递，非指针传递。 

参数数量均不能超过 5 个。 

多返回值最多返回三个，超过三个请使用 struct。

#### defer

当存在资源创建时，应紧跟 defer 释放资源 (可以大胆使用 defer，defer 在 Go1.14 版本中，性能大幅提升，defer 的性能损耗即使在性能敏感型的业务中，也可以忽略)。 

先判断是否错误，再 defer 释放资源，例如：

```go
resp, err := http.Get(url)
if err != nil {
	return err
}

defer resp.Body.Close()
```

#### 方法的接收器

推荐以类名第一个英文首字母的小写作为接收器的命名。 

接收器的命名在函数超过 20 行的时候不要用单字符。 

接收器的命名不能采用 me、this、self 这类易混淆名称。

#### 嵌套

嵌套深度不能超过 4 层。

#### 变量命名

变量声明尽量放在变量第一次使用的前面，遵循就近原则。 

如果魔法数字出现超过两次，则禁止使用，改用一个常量代替，例如：

```go
// PI ...
const Prise = 3.14

func getAppleCost(n float64) float64 {
	return Prise * n
}

func getOrangeCost(n float64) float64 {
	return Prise * n
}
```

### GOPATH 设置规范

Go 1.11 之后，弱化了 GOPATH 规则，已有代码（很多库肯定是在 1.11 之前建立的） 肯定符合这个规则，建议保留 GOPATH 规则，便于维护代码。 

建议只使用一个 GOPATH，不建议使用多个 GOPATH。如果使用多个 GOPATH，编译生效的 bin 目录是在第一个 GOPATH 下。

### 依赖管理

Go 1.11 以上必须使用 Go Modules。 

使用 Go Modules 作为依赖管理的项目时，不建议提交 vendor 目录。 

使用 Go Modules 作为依赖管理的项目时，必须提交 go.sum 文件。

### 最佳实践

尽量少用全局变量，而是通过参数传递，使每个函数都是“无状态”的。这样可以减少耦合，也方便分工和单元测试。 

在编译时验证接口的符合性，例如：

```go
type LogHandler struct {
  h http.Handler
  log *zap.Logger
}

var _ http.Handler = LogHandler{}
```

服务器处理请求时，应该创建一个 context，保存该请求的相关信息（如 requestID），并在函数调用链中传递。

#### 性能

string 表示的是不可变的字符串变量，对 string 的修改是比较重的操作，基本上都需要重新申请内存。所以，如果没有特殊需要，需要修改时多使用 []byte。 

优先使用 strconv 而不是 fmt。

#### 注意事项

append 要小心自动分配内存，append 返回的可能是新分配的地址。 

如果要直接修改 map 的 value 值，则 value 只能是指针，否则要覆盖原来的值。 

map 在并发中需要加锁。 编译过程无法检查 interface{} 的转换，只能在运行时检查，小心引起 panic。

### 总结

介绍了九类常用的编码规范。

要提醒一句：规范是人定的，也可以根据需要，制定符合项目的规范。但同时也建议采纳这些业界沉淀下来的规范，并通过工具来确保规范的执行。



## API 风格设计之 RESTful API

如何设计应用的 API 风格。 

绝大部分的 Go 后端服务需要编写 API 接口，对外提供服务。所以在开发之前，需要确定一种 API 风格。

API 风格也可以理解为 API 类型，目前业界常用的 API 风格有三种： REST、RPC 和 GraphQL。

需要根据项目需求，并结合 API 风格的特点，确定使用哪种 API 风格，这对以后的编码实现、通信方式和通信效率都有很大的影响。 

在 Go 项目开发中，用得最多的是 REST 和 RPC，在 IAM 实战项目中也使用了 REST 和 RPC 来构建示例项目。

接下来会详细介绍下 REST 和 RPC 这两种风格，如果对 GraphQL 感兴趣，GraphQL 中文官网有很多文档和代码示例，可以自行学习。

### RESTful API 介绍 

在回答“RESTful API 是什么”之前，先来看下 REST 是什么意思：REST 代表的是表 现层状态转移（REpresentational State Transfer），由 Roy Fielding 在他的论文 《Architectural Styles and the Design of Network-based Software Architectures》里提出。

REST 本身并没有创造新的技术、组件或服务，它只是一种软件架构风格，是一组架构约束条件和原则，而不是技术框架。 

REST 有一系列规范，满足这些规范的 API 均可称为 RESTful API。REST 规范把所有内容都视为资源，也就是说网络上一切皆资源。REST 架构对资源的操作包括获取、创建、修改和删除，这些操作正好对应 HTTP 协议提供的 GET、POST、PUT 和 DELETE 方法。 

HTTP 动词与 REST 风格 CRUD 的对应关系见下表：

![image-20211115212612806](IAM-document.assets/image-20211115212612806.png)

REST 风格虽然适用于很多传输协议，但在实际开发中，由于 REST 天生和 HTTP 协议相辅相成，因此 HTTP 协议已经成了实现 RESTful API 事实上的标准。所以，REST 具有以下核心特点：

- 以资源 (resource) 为中心，所有的东西都抽象成资源，所有的行为都应该是在资源上的 CRUD 操作。 
  - 资源对应着面向对象范式里的对象，面向对象范式以对象为中心。 
  - 资源使用 URI 标识，每个资源实例都有一个唯一的 URI 标识。例如，如果有一个用户，用户名是 admin，那么它的 URI 标识就可以是 /users/admin。 
- 资源是有状态的，使用 JSON/XML 等在 HTTP Body 里表征资源的状态。
- 客户端通过四个 HTTP 动词，对服务器端资源进行操作，实现“表现层状态转化”。 
- 无状态，这里的无状态是指每个 RESTful API 请求都包含了所有足够完成本次操作的信息，服务器端无须保持 session。无状态对于服务端的弹性扩容是很重要的。

这里强调下 REST 和 RESTful API 的区别：REST 是一种规范，而 RESTful API 则是满足这种规范的 API 接口。

### RESTful API 设计原则 

RESTful API 就是满足 REST 规范的 API，由此看来，RESTful API 的核心是规范，那么具体有哪些规范呢？ 

接下来，就从 URI 设计、API 版本管理等七个方面，详细介绍下 RESTful API 的设 计原则，然后再通过一个示例来快速启动一个 RESTful API 服务。

#### URI 设计 

资源都是使用 URI 标识的，应该按照一定的规范来设计 URI，通过规范化可以使 API 接口更加易读、易用。以下是 URI 设计时，应该遵循的一些规范：

- 资源名使用名词而不是动词，并且用名词复数表示。资源分为 Collection 和 Member 两种。 

  - Collection：一堆资源的集合。例如系统里有很多用户（User）, 这些用户的集合就是 Collection。Collection 的 URI 标识应该是 域名/资源名复数, 例如 https:// iam.api.marmotedu.com/users。 
  - Member：单个特定资源。例如系统中特定名字的用户，就是 Collection 里的一个 Member。Member 的 URI 标识应该是 域名/资源名复数/资源名称, 例如 https:// iam.api.marmotedu/users/admin。 

- URI 结尾不应包含/。 

- URI 中不能出现下划线 _，必须用中杠线 -代替（有些人推荐用 _，有些人推荐用 -，统一使用一种格式即可，比较推荐用 -）。 

- URI 路径用小写，不要用大写。

- 避免层级过深的 URI。超过 2 层的资源嵌套会很乱，建议将其他资源转化为?参数，比如：

  - ```http
    /schools/tsinghua/classes/rooma/students/zhang # 不推荐
    /students?school=qinghua&class=rooma # 推荐
    ```

  - 

这里有个地方需要注意：在实际的 API 开发中，可能会发现有些操作不能很好地映射为一个 REST 资源，这时候，可以参考下面的做法。

- 将一个操作变成资源的一个属性，比如想在系统中暂时禁用某个用户，可以这么设计 URI：/users/zhangsan?active=false。 

- 将操作当作是一个资源的嵌套资源，比如一个 GitHub 的加星操作：

  - ```http
    PUT /gists/:id/star # github star action
    DELETE /gists/:id/star # github unstar action
    ```

  - 如果以上都不能解决问题，有时可以打破这类规范。比如登录操作，登录不属于任何一个资源，URI 可以设计为：/login。

在设计 URI 时，如果遇到一些不确定的地方，推荐参考 GitHub 标准 RESTful API。

#### REST 资源操作映射为 HTTP 方法 

基本上 RESTful API 都是使用 HTTP 协议原生的 GET、PUT、POST、DELETE 来标识对资源的 CRUD 操作的，形成的规范如下表所示：

![image-20211115213736675](IAM-document.assets/image-20211115213736675.png)

对资源的操作应该满足安全性和幂等性：

- 安全性：不会改变资源状态，可以理解为只读的。 
- 幂等性：执行 1 次和执行 N 次，对资源状态改变的效果是等价的。

使用不同 HTTP 方法时，资源操作的安全性和幂等性对照见下表：

![image-20211115213911384](IAM-document.assets/image-20211115213911384.png)

在使用 HTTP 方法的时候，有以下两点需要注意：

- GET 返回的结果，要尽量可用于 PUT、POST 操作中。例如，用 GET 方法获得了一个 user 的信息，调用者修改 user 的邮件，然后将此结果再用 PUT 方法更新。这要求 GET、PUT、POST 操作的资源属性是一致的。 
- 如果对资源进行状态 / 属性变更，要用 PUT 方法，POST 方法仅用来创建或者批量删除这两种场景。

在设计 API 时，经常会有批量删除的需求，需要在请求中携带多个需要删除的资源名，但是 HTTP 的 DELETE 方法不能携带多个资源名，这时候可以通过下面三种方式来解决：

- 发起多个 DELETE 请求。 
- 操作路径中带多个 id，id 之间用分隔符分隔, 例如：DELETE /users?ids=1,2,3 。 
- 直接使用 POST 方式来批量删除，body 中传入需要删除的资源列表。

其中，第二种是最推荐的方式，因为使用了匹配的 DELETE 动词，并且不需要发送多次 DELETE 请求。 

需要注意的是，这三种方式都有各自的使用场景，可以根据需要自行选择。如果选择了某一种方式，那么整个项目都需要统一用这种方式。

#### 统一的返回格式 

一般来说，一个系统的 RESTful API 会向外界开放多个资源的接口，每个接口的返回格式要保持一致。

另外，每个接口都会返回成功和失败两种消息，这两种消息的格式也要保持一致。不然，客户端代码要适配不同接口的返回格式，每个返回格式又要适配成功和失败 两种消息格式，会大大增加用户的学习和使用成本。 

返回的格式没有强制的标准，可以根据实际的业务需要返回不同的格式。后续内容中，会推荐一种返回格式，它也是业界最常用和推荐的返回格式。

#### API 版本管理 

随着时间的推移、需求的变更，一个 API 往往满足不了现有的需求，这时候就需要对 API 进行修改。对 API 进行修改时，不能影响其他调用系统的正常使用，这就要求 API 变更做到向下兼容，也就是新老版本共存。 

但在实际场景中，很可能会出现同一个 API 无法向下兼容的情况。这时候最好的解决办法是从一开始就引入 API 版本机制，当不能向下兼容时，就引入一个新的版本，老的版本则保留原样。这样既能保证服务的可用性和安全性，同时也能满足新需求。

API 版本有不同的标识方法，在 RESTful API 开发中，通常将版本标识放在如下 3 个位置：

- URL 中，比如/v1/users。 
- HTTP Header 中，比如Accept: vnd.example-com.foo+json; version=1.0。 
- Form 参数中，比如/users?version=v1。

这门课中的版本标识是放在 URL 中的，比如/v1/users，这样做的好处是很直观， GitHub、Kubernetes、Etcd 等很多优秀的 API 均采用这种方式。 

这里要注意，有些开发人员不建议将版本放在 URL 中，因为他们觉得不同的版本可以理解成同一种资源的不同表现形式，所以应该采用同一个 URI。对于这一点，没有严格的标准，根据项目实际需要选择一种方式即可。

#### API 命名 

API 通常的命名方式有三种，分别是驼峰命名法 (serverAddress)、蛇形命名法 (server_address) 和脊柱命名法 (server-address)。 

驼峰命名法和蛇形命名法都需要切换输入法，会增加操作的复杂性，也容易出错，所以这里建议用脊柱命名法。GitHub API 用的就是脊柱命名法，例如  selected-actions。

#### 统一分页 / 过滤 / 排序 / 搜索功能 

REST 资源的查询接口，通常情况下都需要实现分页、过滤、排序、搜索功能，因为这些功能是每个 REST 资源都能用到的，所以可以实现为一个公共的 API 组件。

下面来介绍下这些功能。

- 分页：在列出一个 Collection 下所有的 Member 时，应该提供分页功能，例 如/users?offset=0&limit=20（limit，指定返回记录的数量；offset，指定返回记录的开始位置）。引入分页功能可以减少 API 响应的延时，同时可以避免返回太多条目，导致服务器 / 客户端响应特别慢，甚至导致服务器 / 客户端 crash 的情况。 
- 过滤：如果用户不需要一个资源的全部状态属性，可以在 URI 参数里指定返回哪些属性，例如/users?fields=email,username,address。 
- 排序：用户很多时候会根据创建时间或者其他因素，列出一个 Collection 中前 100 个 Member，这时可以在 URI 参数中指明排序参数，例如/users?sort=age,desc。
- 搜索：当一个资源的 Member 太多时，用户可能想通过搜索，快速找到所需要的 Member，或着想搜下有没有名字为 xxx 的某类资源，这时候就需要提供搜索功能。搜索建议按模糊匹配来搜索。

#### 域名 

API 的域名设置主要有两种方式：

- https://marmotedu.com/api ，这种方式适合 API 将来不会有进一步扩展的情况， 比如刚开始 marmotedu.com 域名下只有一套 API 系统，未来也只有这一套 API 系统。 
- https://iam.api.marmotedu.com，如果 marmotedu.com 域名下未来会新增另一个系统 API，这时候最好的方式是每个系统的 API 拥有专有的 API 域名，比如： storage.api.marmotedu.com，network.api.marmotedu.com。腾讯云的域名就是采用这种方式。

到这里，就将 REST 设计原则中的核心原则讲完了，这里有个需要注意的点：不同公司、不同团队、不同项目可能采取不同的 REST 设计原则，以上所列的基本上都是大家公认的原则。 

REST 设计原则中，还有一些原则因为内容比较多，并且可以独立成模块，所以放在后面来讲。比如 RESTful API 安全性、状态返回码和认证等。

### REST 示例 

上面介绍了一些概念和原则，这里通过一个“Hello World”程序，来用 Go 快速启动一个 RESTful API 服务，示例代码存放在gopractisedemo/apistyle/ping/main.go。

```go
package main

import (
   "log"
   "net/http"
)

func main() {
   http.HandleFunc("/ping", pong)
   log.Println("Starting http server ...")
   log.Fatal(http.ListenAndServe(":50052", nil))
}

func pong(w http.ResponseWriter, r *http.Request) {
   w.Write([]byte("pong"))
}
```

在上面的代码中，通过 http.HandleFunc，向 HTTP 服务注册了一个 pong handler，在 pong handler 中，编写了真实的业务代码：返回 pong 字符串。 

创建完 main.go 文件后，在当前目录下执行 go run main.go 启动 HTTP 服务，在一个新的 Linux 终端下发送 HTTP 请求，进行使用 curl 命令测试：

```go
$ curl http://127.0.0.1:50052/ping
pong
```

### 总结 

介绍了两种常用 API 风格中的一种，RESTful API。

REST 是一种 API 规范，而 RESTful API 则是满足这种规范的 API 接口，RESTful API 的核心是规范。 

在 REST 规范中，资源通过 URI 来标识，资源名使用名词而不是动词，并且用名词复数表示，资源都是分为 Collection 和 Member 两种。

RESTful API 中，分别使用 POST、 DELETE、PUT、GET 来表示 REST 资源的增删改查，HTTP 方法、Collection、Member 不同组合会产生不同的操作，具体的映射可以看下 REST 资源操作映射为 HTTP 方法部分的表格。 

为了方便用户使用和理解，每个 RESTful API 的返回格式、错误和正确消息的返回格式，都应该保持一致。

RESTful API 需要支持 API 版本，并且版本应该能够向前兼容，可以将版本号放在 URL 中、HTTP Header 中、Form 参数中，但这里建议将版本号放在 URL 中，例如 /v1/users，这种形式比较直观。

另外，可以通过脊柱命名法来命名 API 接口名。对于一个 REST 资源，其查询接口还应该支持分页 / 过滤 / 排序 / 搜索功能，这些功能可以用同一套机制来实现。 

API 的域名可以采用 https://marmotedu.com/api 和 https://iam.api.marmotedu.com 两种格式。 

最后，在 Go 中可以使用 net/http 包来快速启动一个 RESTful API 服务。 

### 课后练习

- 使用 net/http 包，快速实现一个 RESTful API 服务，并实现 /hello 接口，该接口会返回“Hello World”字符串。 
- 思考一下，RESTful API 这种 API 风格是否能够满足当前的项目需要，如果不满足， 原因是什么？



## API 风格设计之 RPC API

如何设计应用的 API 风格。 

上一讲，介绍了 REST API 风格，这一讲来介绍下另外一种常用的 API 风格，RPC。 

在 Go 项目开发中，如果业务对性能要求比较高，并且需要提供给多种编程语言调用，这时候就可以考虑使用 RPC API 接口。RPC 在 Go 项目开发中用得也非常多，需要认真掌握。 

### RPC 介绍 

根据维基百科的定义，RPC（Remote Procedure Call），即远程过程调用，是一个计算机通信协议。该协议允许运行于一台计算机的程序调用另一台计算机的子程序，而程序员不用额外地为这个交互作用编程。

通俗来讲，就是服务端实现了一个函数，客户端使用 RPC 框架提供的接口，像调用本地函数一样调用这个函数，并获取返回值。RPC 屏蔽了底层的网络通信细节，使得开发人员无需关注网络编程的细节，可以将更多的时间和精力放在业务逻辑本身的实现上，从而提高开发效率。 

RPC 的调用过程如下图所示：

![image-20211115223125427](IAM-document.assets/image-20211115223125427.png)

RPC 调用具体流程如下：

1. Client 通过本地调用，调用 Client Stub。 
2. Client Stub 将参数打包（也叫 Marshalling）成一个消息，然后发送这个消息。 
3. Client 所在的 OS 将消息发送给 Server。
4. Server 端接收到消息后，将消息传递给 Server Stub。
5. Server Stub 将消息解包（也叫 Unmarshalling）得到参数。 
6. Server Stub 调用服务端的子程序（函数），处理完后，将最终结果按照相反的步骤返回给 Client。

这里需要注意，Stub 负责调用参数和返回值的流化（serialization）、参数的打包和解包，以及网络层的通信。Client 端一般叫 Stub，Server 端一般叫 Skeleton。 

目前，业界有很多优秀的 RPC 协议，例如腾讯的 Tars、阿里的 Dubbo、微博的 Motan、Facebook 的 Thrift、RPCX，等等。但使用最多的还是 gRPC，这也是本专栏所采用的 RPC 框架，所以接下来会重点介绍 gRPC 框架。

### gRPC 介绍 

gRPC 是由 Google 开发的高性能、开源、跨多种编程语言的通用 RPC 框架，基于 HTTP 2.0 协议开发，默认采用 Protocol Buffers 数据序列化协议。gRPC 具有如下特性：

- 支持多种语言，例如 Go、Java、C、C++、C#、Node.js、PHP、Python、Ruby 等。 
- 基于 IDL（Interface Definition Language）文件定义服务，通过 proto3 工具生成指定语言的数据结构、服务端接口以及客户端 Stub。通过这种方式，也可以将服务端和客户端解耦，使客户端和服务端可以并行开发。 
- 通信协议基于标准的 HTTP/2 设计，支持双向流、消息头压缩、单 TCP 的多路复用、服务端推送等特性。 
- 支持 Protobuf 和 JSON 序列化数据格式。Protobuf 是一种语言无关的高性能序列化框架，可以减少网络传输流量，提高通信效率。

这里要注意的是，gRPC 的全称不是 golang Remote Procedure Call，而是 google Remote Procedure Call。 

gRPC 的调用如下图所示：

![image-20211115223619830](IAM-document.assets/image-20211115223619830.png)

在 gRPC 中，客户端可以直接调用部署在不同机器上的 gRPC 服务所提供的方法，调用远端的 gRPC 方法就像调用本地的方法一样，非常简单方便，通过 gRPC 调用，可以非常容易地构建出一个分布式应用。 

像很多其他的 RPC 服务一样，gRPC 也是通过 IDL 语言，预先定义好接口（接口的名字、 传入参数和返回参数等）。在服务端，gRPC 服务实现所定义的接口。在客户端， gRPC 存根提供了跟服务端相同的方法。 

gRPC 支持多种语言，比如可以用 Go 语言实现 gRPC 服务，并通过 Java 语言客户端调用 gRPC 服务所提供的方法。通过多语言支持，编写的 gRPC 服务能满足客户端多语言的需求。 

gRPC API 接口通常使用的数据传输格式是 Protocol Buffers。接下来，就一起了解下 Protocol Buffers。

### Protocol Buffers 介绍 

Protocol Buffers（ProtocolBuffer/ protobuf）是 Google 开发的一套对数据结构进行序列化的方法，可用作（数据）通信协议、数据存储格式等，也是一种更加灵活、高效的数据格式，与 XML、JSON 类似。它的传输性能非常好，所以常被用在一些对数据传输性能要求比较高的系统中，作为数据传输格式。Protocol Buffers 的主要特性有下面这几个。

- 更快的数据传输速度：protobuf 在传输时，会将数据序列化为二进制数据，和 XML、 JSON 的文本传输格式相比，这可以节省大量的 IO 操作，从而提高数据传输速度。 
- 跨平台多语言：protobuf 自带的编译工具 protoc 可以基于 protobuf 定义文件，编译出不同语言的客户端或者服务端，供程序直接调用，因此可以满足多语言需求的场景。 
- 具有非常好的扩展性和兼容性，可以更新已有的数据结构，而不破坏和影响原有的程序。 
- 基于 IDL 文件定义服务，通过 proto3 工具生成指定语言的数据结构、服务端和客户端接口。

在 gRPC 的框架中，Protocol Buffers 主要有三个作用。 

第一，可以用来定义数据结构。

举个例子，下面的代码定义了一个 SecretInfo 数据结构：

```go
// SecretInfo contains secret details.
message SecretInfo {
  string name = 1;
  string secret_id = 2;
  string username = 3;
  string secret_key = 4;
  int64 expires = 5;
  string description = 6;
  string created_at = 7;
  string updated_at = 8;
}
```

第二，可以用来定义服务接口。

下面的代码定义了一个 Cache 服务，服务包含了 ListSecrets 和 ListPolicies 两个 API 接口。

```go
// Cache implements a cache rpc service.
service Cache{
rpc ListSecrets(ListSecretsRequest) returns (ListSecretsResponse) {}
rpc ListPolicies(ListPoliciesRequest) returns (ListPoliciesResponse) {}
}
```

第三，可以通过 protobuf 序列化和反序列化，提升传输效率。

### gRPC 示例 

已经对 gRPC 这一通用 RPC 框架有了一定的了解，可能还不清楚怎么使用 gRPC 编写 API 接口。接下来，就通过 gRPC 官方的一个示例来快速给大家展示下。运行本示例需要在 Linux 服务器上安装 Go 编译器、Protocol buffer 编译器（protoc， v3）和 protoc 的 Go 语言插件。

这个示例分为下面几个步骤：

- 定义 gRPC 服务。 
- 生成客户端和服务器代码。 
- 实现 gRPC 服务。 
- 实现 gRPC 客户端。

示例代码存放在 gopractise-demo/apistyle/greeter目录下。代码结构如下：

```sh
$ tree
.
├── client
│   └── main.go
├── helloworld
│   ├── helloworld.pb.go
│   └── helloworld.proto
└── server
    └── main.go
```

client 目录存放 Client 端的代码，helloworld 目录用来存放服务的 IDL 定义，server 目录用来存放 Server 端的代码。 

下面具体介绍下这个示例的四个步骤。

#### 定义 gRPC 服务

首先，需要定义服务。进入 helloworld 目录，新建文件 helloworld.proto：

```sh
$ cd helloworld
$ vi helloworld.proto
```

内容如下：

```protobuf
syntax = "proto3";

option go_package = "github.com/marmotedu/gopractise-demo/apistyle/greeter/helloworld";

package helloworld;

// The greeting service definition.
service Greeter {
  // Sends a greeting
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}

// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}

// The response message containing the greetings
message HelloReply {
  string message = 1;
}
```

在 helloworld.proto 定义文件中，

- option 关键字用来对.proto 文件进行一些设置，其中 go_package 是必需的设置，而且 go_package 的值必须是包导入的路径。

- package 关键字指定生成的.pb.go 文件所在的包名。

- 通过 service 关键字定义服务，然后再指定该服务拥有的 RPC 方法，并定义方法的请求和返回的结构体类型：

  - ```protobuf
    service Greeter {
      // Sends a greeting
      rpc SayHello (HelloRequest) returns (HelloReply) {}
    }
    ```

gRPC 支持定义 4 种类型的服务方法，分别是简单模式、服务端数据流模式、客户端数据流模式和双向数据流模式。

- 简单模式（Simple RPC）：是最简单的 gRPC 模式。客户端发起一次请求，服务端响应一个数据。定义格式为 rpc SayHello (HelloRequest) returns (HelloReply) {}。 
- 服务端数据流模式（Server-side streaming RPC）：客户端发送一个请求，服务器返回数据流响应，客户端从流中读取数据直到为空。定义格式为 rpc SayHello (HelloRequest) returns (stream HelloReply) {}。 
- 客户端数据流模式（Client-side streaming RPC）：客户端将消息以流的方式发送给服务器，服务器全部处理完成之后返回一次响应。定义格式为 rpc SayHello (stream HelloRequest) returns (HelloReply) {}。 
- 双向数据流模式（Bidirectional streaming RPC）：客户端和服务端都可以向对方发送数据流，这个时候双方的数据可以同时互相发送，也就是可以实现实时交互 RPC 框架原理。定义格式为 rpc SayHello (stream HelloRequest) returns (stream HelloReply) {}。

本示例使用了简单模式。.proto 文件也包含了 Protocol Buffers 消息的定义，包括请求消息和返回消息。例如请求消息：

```protobuf
// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}
```

#### 生成客户端和服务器代码

接下来，需要根据.proto 服务定义生成 gRPC 客户端和服务器接口。可以使用 protoc 编译工具，并指定使用其 Go 语言插件来生成：

```sh
$ protoc -I. --go_out=plugins=grpc:$GOPATH/src helloworld.proto
$ ls
helloworld.pb.go helloworld.proto
```

可以看到，新增了一个 helloworld.pb.go 文件。

#### 实现 gRPC 服务

接着，就可以实现 gRPC 服务了。进入 server 目录，新建 main.go 文件：

```sh
$ cd ../server
$ vi main.go
```

main.go 内容如下：

```go
package main

import (
	"context"
	"log"
	"net"

	pb "github.com/marmotedu/gopractise-demo/apistyle/greeter/helloworld"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
```

上面的代码实现了上一步根据服务定义生成的 Go 接口。 

先定义了一个 Go 结构体 server，并为 server 结构体添加 SayHello(context.Context, pb.HelloRequest) (pb.HelloReply, error) 方法，也就是说 server 是 GreeterServer 接口（位于 helloworld.pb.go 文件中）的一个实现。 

在实现了 gRPC 服务所定义的方法之后，就可以通过 net.Listen(...) 指定监听客户端请求的端口；接着，通过 grpc.NewServer() 创建一个 gRPC Server 实例，并通过 pb.RegisterGreeterServer(s, &server{}) 将该服务注册到 gRPC 框架中；最后，通过 s.Serve(lis) 启动 gRPC 服务。 

创建完 main.go 文件后，在当前目录下执行 `go run main.go` ，启动 gRPC 服务。

#### 实现 gRPC 客户端

打开一个新的 Linux 终端，进入 client 目录，新建 main.go 文件：

```sh
$ cd ../client
$ vi main.go
```

main.go 内容如下：

```go
// Package main implements a client for Greeter service.
package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/marmotedu/gopractise-demo/apistyle/greeter/helloworld"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
```

在上面的代码中，通过如下代码创建了一个 gRPC 连接，用来跟服务端进行通信：

```go
// Set up a connection to the server.
conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
if err != nil {
   log.Fatalf("did not connect: %v", err)
}
defer conn.Close()
```

在创建连接时，可以指定不同的选项，用来控制创建连接的方式，例如 grpc.WithInsecure()、grpc.WithBlock() 等。gRPC 支持很多选项，更多的选项可以参考 grpc 仓库下 dialoptions.go 文件中以 With 开头的函数。 

连接建立起来之后，需要创建一个客户端 stub，用来执行 RPC 请求c := pb.NewGreeterClient(conn)。

创建完成之后，就可以像调用本地函数一样，调用远程的方法了。例如，下面一段代码通过 c.SayHello 这种本地式调用方式调用了远端 的 SayHello 接口：

```go
r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
if err != nil {
   log.Fatalf("could not greet: %v", err)
}
log.Printf("Greeting: %s", r.Message)
```

从上面的调用格式中，可以看到 RPC 调用具有下面两个特点。

- 调用方便：RPC 屏蔽了底层的网络通信细节，使得调用 RPC 就像调用本地方法一样方便，调用方式跟大家所熟知的调用类的方法一致：ClassName.ClassFuc(params)。 
- 不需要打包和解包：RPC 调用的入参和返回的结果都是 Go 的结构体，不需要对传入参数进行打包操作，也不需要对返回参数进行解包操作，简化了调用步骤。

最后，创建完 main.go 文件后，在当前目录下，执行 go run main.go 发起 RPC 调用：

```sh
$ go run main.go
2021/11/15 23:57:42 Greeting: Hello world

# 如果遇到错误，eg: This download does NOT match an earlier download recorded in go.sum.
# 解决办法：
# remove go.sum : rm go.sum
# regenerate go.sum : go mod tidy
```

至此，用四个步骤，创建并调用了一个 gRPC 服务。

### 具体场景：指针判断 nil

接下来再讲解一个在具体场景中的注意事项。 

在做服务开发时，经常会遇到一种场景：定义一个接口，接口会通过判断是否传入某个参数，决定接口行为。例如，想提供一个 GetUser 接口，期望 GetUser 接口在传入 username 参数时，根据 username 查询用户的信息，如果没有传入 username，则默认根据 userId 查询用户信息。 

这时候，需要判断客户端有没有传入 username 参数。不能根据 username 是否 为空值来判断，因为不能区分客户端传的是空值，还是没有传 username 参数。这是由 Go 语言的语法特性决定的：如果客户端没有传入 username 参数，Go 会默认赋值为所在类型的零值，而字符串类型的零值就是空字符串。 

那怎么判断客户端有没有传入 username 参数呢？最好的方法是通过指针来判断，如果是 nil 指针就说明没有传入，非 nil 指针就说明传入，具体实现步骤如下：

#### 编写 protobuf 定义文件

新建 user.proto 文件，内容如下:

```protobuf
syntax = "proto3";

package proto;
option go_package = "github.com/marmotedu/gopractise-demo/protobuf/user";

//go:generate protoc -I. --experimental_allow_proto3_optional --go_out=plugins=grpc:. user.proto

service User {
	rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
}

message GetUserRequest {
  string class = 1;
  optional string username = 2;
  optional string user_id = 3;
}

message GetUserResponse {
  string class = 1;
  string user_id = 2;
  string username = 3;
  string address = 4;
  string sex = 5;
  string phone = 6;
}
```

需要注意，这里在需要设置为可选字段的前面添加了 optional 标识。

#### 使用 protoc 工具编译 protobuf 文件

在执行 protoc 命令时，需要传入--experimental_allow_proto3_optional参数以打开 optional 选项，编译命令如下：

```sh
$ protoc --experimental_allow_proto3_optional --go_out=plugins=grpc:. user.proto
```

上述编译命令会生成 user.pb.go 文件，其中的 GetUserRequest 结构体定义如下：

```go
type GetUserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Class    string `protobuf:"bytes,1,opt,name=class,proto3" json:"class,omitempty"`
	UserId   string `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Username string `protobuf:"bytes,3,opt,name=username,proto3" json:"username,omitempty"`
}
```

通过 optional + --experimental_allow_proto3_optional 组合，可以将一个字段编译为指针类型。

#### 编写 gRPC 接口实现

新建一个 user.go 文件，内容如下：

```go
package user

import (
	"context"

	pb "github.com/marmotedu/api/proto/apiserver/v1"

	"github.com/marmotedu/iam/internal/apiserver/store"
)

type User struct {
}

func (c *User) GetUser(ctx context.Context, r *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if r.Username != nil {
		return store.Client().Users().GetUserByName(r.Class, r.Username)
	}

	return store.Client().Users().GetUserByID(r.Class, r.UserId)
}
```

总之，在 GetUser 方法中，可以通过判断 r.Username 是否为 nil，来判断客户端是否传入了 Username 参数。

### RESTful VS gRPC 

到这里，已经介绍完了 gRPC API。回想一下 RESTful API，可能想问：这两种 API 风格分别有什么优缺点，适用于什么场景呢？把这个问题的答案放在了下面这张表中，可以对照着它，根据自己的需求在实际应用时进行选择。

![image-20211116002601073](IAM-document.assets/image-20211116002601073.png)

当然，更多的时候，RESTful API 和 gRPC API 是一种合作的关系，对内业务使用 gRPC API，对外业务使用 RESTful API，如下图所示：

![image-20211116002803542](IAM-document.assets/image-20211116002803542.png)

### 总结 

在 Go 项目开发中，可以选择使用 RESTful API 风格和 RPC API 风格，这两种服务都用得很多。

- 其中，RESTful API 风格因为规范、易理解、易用，所以适合用在需要对外提供 API 接口的场景中。
- 而 RPC API 因为性能比较高、调用方便，更适合用在内部业务中。 
- RESTful API 使用的是 HTTP 协议，而 RPC API 使用的是 RPC 协议。

目前，有很多 RPC 协议可供选择，推荐使用 gRPC，因为它很轻量，同时性能很高、很稳定，是一个优秀的 RPC 框架。所以目前业界用的最多的还是 gRPC 协议，腾讯、阿里等大厂内部很多核心的线上服务用的就是 gRPC。 

除了使用 gRPC 协议，在进行 Go 项目开发前，也可以了解业界一些其他的优秀 Go RPC 框架，比如腾讯的 tars-go、阿里的 dubbo-go、Facebook 的 thrift、rpcx 等，可以在项目开发之前一并调研，根据实际情况进行选择。

### 课后练习

- 使用 gRPC 包，快速实现一个 RPC API 服务，并实现 PrintHello 接口，该接口会返回“Hello World”字符串。 
- 请思考这个场景：有一个 gRPC 服务，但是却希望该服务同时也能提供 RESTful API 接口，这该如何实现？
  - 假定希望用RPC作为内部API的通讯，同时也想对外提供RESTful API，又不想写两套，可以使用gRPC Gateway 插件，在生成RPC的同时也生成RESTful web server。



## Go项目管理之 Makefile

如何编写高质量的 Makefile。 

要写出一个优雅的 Go 项目，不仅仅是要开发一个优秀的 Go 应用，而且还要能够高效地管理项目。有效手段之一，就是通过 Makefile 来管理项 目，这就要求为项目编写 Makefile 文件。 

在和其他开发同学交流时，发现大家都认可 Makefile 强大的项目管理能力，也会自己编写 Makefile。但是其中的一些人项目管理做得并不好，和他们进一步交流后发现，这些同学在用 Makefile 简单的语法重复编写一些低质量 Makefile 文件，根本没有把 Makefile 的功能充分发挥出来。 

下面举个例子，就会理解低质量的 Makefile 文件是什么样的了。

```makefile
build: clean vet
  @mkdir -p ./Role
  @export GOOS=linux && go build -v .

vet:
  go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf dashboard
```

上面这个 Makefile 存在不少问题。例如：

- 功能简单，只能完成最基本的编译、格式化等操作，像构建镜像、自动生成代码等一些高阶的功能都没有；
- 扩展性差，没法编译出可在 Mac 下运行的二进制文件；
- 没有 Help 功能，使用难度高；
- 单 Makefile 文件，结构单一，不适合添加一些复杂的管理功能。 

所以，不光要编写 Makefile，还要编写高质量的 Makefile。那么如何编写一个高质量的 Makefile 呢？可以通过以下 4 个方法来实现：

- 打好基础，也就是熟练掌握 Makefile 的语法。
- 做好准备工作，也就是提前规划 Makefile 要实现的功能。 
- 进行规划，设计一个合理的 Makefile 结构。 
- 掌握方法，用好 Makefile 的编写技巧。

### 熟练掌握 Makefile 语法

工欲善其事，必先利其器。编写高质量 Makefile 的第一步，便是熟练掌握 Makefile 的核心语法。 

因为 Makefile 的语法比较多，把一些建议重点掌握的语法放在了近期会更新的特别放送中，包括 Makefile 规则语法、伪目标、变量赋值、条件语句和 Makefile 常用函数等等。

如果想更深入、全面地学习 Makefile 的语法，推荐学习陈皓老师编写的《跟我一 起写 Makefile》 (PDF 重制版)。

### 规划 Makefile 要实现的功能 

接着，需要规划 Makefile 要实现的功能。提前规划好功能，有利于设计 Makefile 的整体结构和实现方法。 

不同项目拥有不同的 Makefile 功能，这些功能中一小部分是通过目标文件来实现的，但更多的功能是通过伪目标来实现的。对于 Go 项目来说，虽然不同项目集成的功能不一样， 但绝大部分项目都需要实现一些通用的功能。接下来，就来看看，在一个大型 Go 项目中 Makefile 通常可以实现的功能。 

下面是 IAM 项目的 Makefile 所集成的功能，希望会对日后设计 Makefile 有一些帮 助。

```sh
$ make help

Usage: make <TARGETS> <OPTIONS> ...

Targets:
  # 代码生成类命令
  gen Generate all necessary files, such as error code files.

  # 格式化类命令
  format Gofmt (reformat) package sources (exclude vendor dir if existed).

  # 静态代码检查
  lint Check syntax and styling of go sources.
  
  # 测试类命令
  test Run unit test.
  cover Run unit test and get test coverage.
  
  # 构建类命令
  build Build source code for host platform.
  build.multiarch Build source code for multiple platforms. See option PLATFORMS.
  
  # Docker镜像打包类命令
  image Build docker images for host arch.
  image.multiarch Build docker images for multiple platforms. See option PLATFORMS.
  push Build docker images for host arch and push images to registry.
  push.multiarch Build docker images for multiple platforms and push images to registry.
  
  # 部署类命令
  deploy Deploy updated components to development env.

  # 清理类命令
  clean Remove all files that are created by building.

  # 其他命令，不同项目会有区别
  release Release iam
  verify-copyright Verify the boilerplate headers for all files.
  ca Generate CA files for all iam components.
  install Install iam system with all its components.
  swagger Generate swagger document.
  tools install dependent tools.
  
  # 帮助命令
  help Show this help info.
  
# 选项
Options:
  DEBUG Whether to generate debug symbols. Default is 0.
  BINS The binaries to build. Default is all of cmd.
  This option is available when using: make build/build.multiarch
  Example: make build BINS="iam-apiserver iam-authz-server"
  ...
```

更详细的命令，可以在 IAM 项目仓库根目录下执行make help查看。 

通常而言，Go 项目的 Makefile 应该实现以下功能：格式化代码、静态代码检查、单元测试、代码构建、文件清理、帮助等等。如果通过 docker 部署，还需要有 docker 镜像打包功能。因为 Go 是跨平台的语言，所以构建和 docker 打包命令，还要能够支持不同的 CPU 架构和平台。为了能够更好地控制 Makefile 命令的行为，还需要支持 Options。 

为了方便查看 Makefile 集成了哪些功能，需要支持 help 命令。help 命令最好通过解析 Makefile 文件来输出集成的功能，例如：

```sh
$ cat Makefile

...
## help: Show this help info.
.PHONY: help
help: Makefile
  @echo -e "\nUsage: make <TARGETS> <OPTIONS> ...\n\nTargets:"
  @sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
  @echo "$$USAGE_OPTIONS"
```

上面的 help 命令，通过解析 Makefile 文件中的##注释，获取支持的命令。通过这种方式，以后新加命令时，就不用再对 help 命令进行修改了。 

可以参考上面的 Makefile 管理功能，结合自己项目的需求，整理出一个 Makefile 要实现的功能列表，并初步确定实现思路和方法。做完这些，编写前准备工作就基本完成了。

### 设计合理的 Makefile 结构 

设计完 Makefile 需要实现的功能，接下来就进入 Makefile 编写阶段。编写阶段的第一步，就是设计一个合理的 Makefile 结构。 

对于大型项目来说，需要管理的内容很多，所有管理功能都集成在一个 Makefile 中，可能会导致 Makefile 很大，难以阅读和维护，所以建议采用分层的设计方法，根目录下的 Makefile 聚合所有的 Makefile 命令，具体实现则按功能分类，放在另外的 Makefile 中。 

经常会在 Makefile 命令中集成 shell 脚本，但如果 shell 脚本过于复杂，也会导致 Makefile 内容过多，难以阅读和维护。并且在 Makefile 中集成复杂的 shell 脚本，编写体验也很差。对于这种情况，可以将复杂的 shell 命令封装在shell 脚本中，供 Makefile 直接调用，而一些简单的命令则可以直接集成在 Makefile 中。 

所以，最终推荐的 Makefile 结构如下：

![image-20211116210829366](IAM-document.assets/image-20211116210829366.png)

在上面的 Makefile 组织方式中，根目录下的 Makefile 聚合了项目所有的管理功能，这些管理功能通过 Makefile 伪目标的方式实现。同时，还将这些伪目标进行分类，把相同类别的伪目标放在同一个 Makefile 中，这样可以使得 Makefile 更容易维护。对于复杂的命令，则编写成独立的 shell 脚本，并在 Makefile 命令中调用这些 shell 脚本。 

举个例子，下面是 IAM 项目的 Makefile 组织结构：

```sh
|-- Makefile                
|-- scripts                 
|		|-- gendoc.sh           
|		|-- make-rules          
|  	|   |-- ca.mk           
|  	|   |-- common.mk       
|  	|   |-- copyright.mk    
|  	|   |-- dependencies.mk 
|  	|   |-- deploy.mk       
|  	|   |-- gen.mk          
|  	|   |-- golang.mk       
|  	|   |-- image.mk        
|  	|   |-- release.mk      
|  	|   |-- swagger.mk      
|  	|   `-- tools.mk        
`-- ...                 
```

将相同类别的操作统一放在 scripts/make-rules 目录下的 Makefile 文件中。 Makefile 的文件名参考分类命名，例如 golang.mk。最后，在 /Makefile 中 include 这些 Makefile。 

为了跟 Makefile 的层级相匹配，golang.mk 中的所有目标都按 go.xxx这种方式命名。通过这种命名方式，可以很容易分辨出某个目标完成什么功能，放在什么文件里，这在复杂的 Makefile 中尤其有用。

以下是 IAM 项目根目录下，Makefile 的内容摘录，可以看一看，作为参考：

```makefile
include scripts/make-rules/common.mk # make sure include common.mk at the first include line
include scripts/make-rules/golang.mk
include scripts/make-rules/image.mk
include scripts/make-rules/gen.mk

## build: Build source code for host platform.
.PHONY: build
build:
	@$(MAKE) go.build

## build.multiarch: Build source code for multiple platforms. See option PLATFORMS.
.PHONY: build.multiarch
build.multiarch:
	@$(MAKE) go.build.multiarch

## image: Build docker images for host arch.
.PHONY: image
image:
	@$(MAKE) image.build

## push: Build docker images for host arch and push images to registry.
.PHONY: push
push:
	@$(MAKE) image.push
	
## ca: Generate CA files for all iam components.
.PHONY: ca
ca:
	@$(MAKE) gen.ca
```

另外，一个合理的 Makefile 结构应该具有前瞻性。也就是说，要在不改变现有结构的情况下，接纳后面的新功能。这就需要整理好 Makefile 当前要实现的功能、即将要实现的功能和未来可能会实现的功能，然后基于这些功能，利用 Makefile 编程技巧，编写可扩展的 Makefile。 

这里需要注意：上面的 Makefile 通过 .PHONY 标识定义了大量的伪目标，定义伪目标一定要加 .PHONY 标识，否则当有同名的文件时，伪目标可能不会被执行。

### 掌握 Makefile 编写技巧 

最后，在编写过程中，还需要掌握一些 Makefile 的编写技巧，这些技巧可以使编写的 Makefile 扩展性更强，功能更强大。 

接下来，会把自己长期开发过程中积累的一些 Makefile 编写经验分享。这些技巧， 需要在实际编写中多加练习，并形成编写习惯。 

#### 技巧 1：善用通配符和自动变量 

Makefile 允许对目标进行类似正则运算的匹配，主要用到的通配符是%。通过使用通配符，可以使不同的目标使用相同的规则，从而使 Makefile 扩展性更强，也更简洁。 

IAM 实战项目中，就大量使用了通配符%，例如：go.build.%、ca.gen.%、 deploy.run.%、tools.verify.%、tools.install.%等。

这里，来看一个具体的例子，tools.verify.%（位于 scripts/make-rules/tools.mk文件中）定义如下：

```makefile
tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi
```

make tools.verify.swagger, make tools.verify.mockgen等均可以使用上面定义的规则，%分别代表了swagger和mockgen。 

如果不使用%，则需要分别为tools.verify.swagger和tools.verify.mockgen 定义规则，很麻烦，后面修改也困难。 

另外，这里也能看出tools.verify.%这种命名方式的好处：tools 说明依赖的定义位于 scripts/make-rules/tools.mk Makefile 中；verify说明tools.verify.%伪目标属于 verify 分类，主要用来验证工具是否安装。通过这种命名方式，可以很容易地知道目标位于哪个 Makefile 文件中，以及想要完成的功能。 

另外，上面的定义中还用到了自动变量$*，用来指代被匹配的值swagger、mockgen。

#### 技巧 2：善用函数 

Makefile 自带的函数能够实现很多强大的功能。所以，在编写 Makefile 的过程中，如果有功能需求，可以优先使用这些函数。把常用的函数以及它们实现的功能 整理在了 Makefile 常用函数列表 中，可以参考下。 

IAM 的 Makefile 文件中大量使用了上述函数，如果想查看这些函数的具体使用方法和场景，可以参考 IAM 项目的 Makefile 文件 make-rules。

#### 技巧 3：依赖需要用到的工具 

如果 Makefile 某个目标的命令中用到了某个工具，可以将该工具放在目标的依赖中。这样，当执行该目标时，就可以指定检查系统是否安装该工具，如果没有安装则自动安装， 从而实现更高程度的自动化。例如，/Makefile 文件中，format 伪目标，定义如下：

```makefile
.PHONY: format
format: tools.verify.golines tools.verify.goimports
	@echo "===========> Formating codes"
	@$(FIND) -type f -name '*.go' | $(XARGS) gofmt -s -w
	@$(FIND) -type f -name '*.go' | $(XARGS) goimports -w -local $(ROOT_PACKAGE)
	@$(FIND) -type f -name '*.go' | $(XARGS) golines -w --max-len=120 --reformat-tags --shorten-comments --ignore-generated .
	@$(GO) mod edit -fmt
```

可以看到，format 依赖tools.verify.golines tools.verify.goimports。再来看下tools.verify.golines的定义：

```makefile
.PHONY: tools.verify.%
tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi
```

再来看下tools.install.$*规则：

```makefile
.PHONY: tools.install.%
tools.install.%:
	@echo "===========> Installing $*"
	@$(MAKE) install.$*
	
.PHONY: install.golines
install.golines:
	@$(GO) install github.com/segmentio/golines@latest
```

通过tools.verify.%规则定义，可以知道，tools.verify.%会先检查工具是否安装，如果没有安装，就会执行tools.install.$*来安装。如此一来，当执行 tools.verify.%目标时，如果系统没有安装 golines 命令，就会自动调用go get安装，提高了 Makefile 的自动化程度。

#### 技巧 4：把常用功能放在 /Makefile 中，不常用的放在分类 Makefile 中 

一个项目，尤其是大型项目，有很多需要管理的地方，其中大部分都可以通过 Makefile 实现自动化操作。不过，为了保持 /Makefile 文件的整洁性，不能把所有的命令都添加在 /Makefile 文件中。 

一个比较好的建议是，将常用功能放在 /Makefile 中，不常用的放在分类 Makefile 中，并在 /Makefile 中 include 这些分类 Makefile。

例如，IAM 项目的 /Makefile 集成了format、lint、test、build等常用命令，而将 gen.errcode.code、gen.errcode.doc这类不常用的功能放在 scripts/make-rules/gen.mk 文件中。当然，也可以直接执行 make gen.errcode.code来执行 make gen.errcode.code伪目标。

通过这种方式，既可以保证 /Makefile 的简洁、易维护，又可以通过make命令来运行伪目标，更加灵活。

#### 技巧 5：编写可扩展的 Makefile 

什么叫可扩展的 Makefile 呢？可扩展的 Makefile 包含两层含义：

- 可以在不改变 Makefile 结构的情况下添加新功能。
- 扩展项目时，新功能可以自动纳入到 Makefile 现有逻辑中。

其中的第一点，可以通过设计合理的 Makefile 结构来实现。要实现第二点，就需要在编写 Makefile 时采用一定的技巧，例如多用通配符、自动变量、函数等。这里来看一个例子，可以更好地理解。 

在 IAM 实战项目的 golang.mk 中，执行 make go.build 时能够构建 cmd/ 目录 下的所有组件，也就是说，当有新组件添加时， make go.build 仍然能够构建新增的组件，这就实现了上面说的第二点。 具体实现方法如下：

```makefile
COMMANDS ?= $(filter-out %.md, $(wildcard ${ROOT_DIR}/cmd/*))
BINS ?= $(foreach cmd,${COMMANDS},$(notdir ${cmd}))

.PHONY: go.build
go.build: go.build.verify $(addprefix go.build., $(addprefix $(PLATFORM)., $(BINS)))

.PHONY: go.build.%
go.build.%:
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	@echo "===========> Building binary $(COMMAND) $(VERSION) for $(OS) $(ARCH)"
	@mkdir -p $(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)
	@CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GO_BUILD_FLAGS) -o $(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)/$(COMMAND)$(GO_OUT_EXT) $(ROOT_PACKAGE)/cmd/$(COMMAND)
```

当执行make go.build 时，会执行 go.build 的依赖 `$(addprefix go.build., $(addprefix $(PLATFORM)., $(BINS))) `,addprefix函数最终返回字符串 go.build.linux_amd64.iamctl go.build.linux_amd64.iam-authz-server go.build.linux_amd64.iam-apiserver ... ，这时候就会执行 go.build.% 伪目标。 

在 go.build.% 伪目标中，通过 eval、word、subst 函数组合，算出了 COMMAND 的值 iamctl/iam-apiserver/iam-authz-server/...，最终通过 `$(ROOT_PACKAGE)/cmd/$(COMMAND) `定位到需要构建的组件的 main 函数所在目录。 

上述实现中有两个技巧，可以注意下。首先，通过

```makefile
COMMANDS ?= $(filter-out %.md, $(wildcard ${ROOT_DIR}/cmd/*))
BINS ?= $(foreach cmd,${COMMANDS},$(notdir ${cmd}))
```

获取到了 cmd/ 目录下的所有组件名。 

接着，通过使用通配符和自动变量，自动匹配到go.build.linux_amd64.iam-authzserver 这类伪目标并构建。 

可以看到，想要编写一个可扩展的 Makefile，熟练掌握 Makefile 的用法是基础，更多的是需要思考如何去编写 Makefile。

#### 技巧 6：将所有输出存放在一个目录下，方便清理和查找 

在执行 Makefile 的过程中，会输出各种各样的文件，例如 Go 编译后的二进制文件、测试覆盖率数据等，建议把这些文件统一放在一个目录下，方便后期的清理和查找。

通常可以把它们放在`_output`这类目录下，这样清理时就很方便，只需要清理`_output`文件夹就可以，例如：

```makefile
.PHONY: go.clean
go.clean:
	@echo "===========> Cleaning all build output"
	@-rm -vrf $(OUTPUT_DIR)
```

这里要注意，要用-rm，而不是rm，防止在没有`_output`目录时，执行make go.clean 报错。

#### 技巧 7：使用带层级的命名方式 

通过使用带层级的命名方式，例如tools.verify.swagger ，可以实现目标分组管 理。这样做的好处有很多。

- 首先，当 Makefile 有大量目标时，通过分组，可以更好地管理这些目标。
- 其次，分组也能方便理解，可以通过组名一眼识别出该目标的功能类别。 
- 最后，这样做还可以大大减小目标重名的概率。 

例如，IAM 项目的 Makefile 就大量采用了下面这种命名方式。

```makefile
.PHONY: gen.run
#gen.run: gen.errcode gen.docgo
gen.run: gen.clean gen.errcode gen.docgo.doc

.PHONY: gen.errcode
gen.errcode: gen.errcode.code gen.errcode.doc

.PHONY: gen.errcode.code
gen.errcode.code: tools.verify.codegen
...

.PHONY: gen.errcode.doc
gen.errcode.doc: tools.verify.codegen
...
```

#### 技巧 8：做好目标拆分 

还有一个比较实用的技巧：要合理地拆分目标。

比如，可以将安装工具拆分成两个目标：验证工具是否已安装和安装工具。通过这种方式，可以给 Makefile 带来更大的灵活性。

例如：可以根据需要选择性地执行其中一个操作，也可以两个操作一起执行。 这里来看一个例子：

```makefile
gen.errcode.code: tools.verify.codegen
	@echo "===========> Generating iam error code go source files"
	@codegen -type=int ${ROOT_DIR}/internal/pkg/code
	
.PHONY: tools.verify.%
tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi
	
.PHONY: install.codegen
install.codegen:
	@$(GO) install ${ROOT_DIR}/tools/codegen/codegen.go
```

上面的 Makefile 中，gen.errcode.code 依赖了 tools.verify.codegen， tools.verify.codegen 会先检查 codegen 命令是否存在，如果不存在，再调用 install.codegen 来安装 codegen 工具。 

如果我们的 Makefile 设计是：

```makefile
gen.errcode.code: install.codegen
```

那每次执行 gen.errcode.code 都要重新安装 codegen 命令，这种操作是不必要的，还会导致 make gen.errcode.code 执行很慢。

#### 技巧 9：设置 OPTIONS 

编写 Makefile 时，还需要把一些可变的功能通过 OPTIONS 来控制。为了帮助理 解，这里还是拿 IAM 项目的 Makefile 来举例。 

假设需要通过一个选项 V ，来控制是否需要在执行 Makefile 时打印详细的信息。这可以通过下面的步骤来实现。 

首先，在 /Makefile 中定义 USAGE_OPTIONS 。定义 USAGE_OPTIONS 可以使开发者在 执行 make help 后感知到此 OPTION，并根据需要进行设置。

```makefile
define USAGE_OPTIONS

Options:
  ...
  BINS         The binaries to build. Default is all of cmd.
               ...
  ...
  V            Set to 1 enable verbose build. Default is 0.
endef
export USAGE_OPTIONS
```

接着，在 scripts/make-rules/common.mk文件中，通过判断有没有设置 V 选项， 来选择不同的行为：

```makefile
ifndef V
MAKEFLAGS += --no-print-directory
endif
```

当然，还可以通过下面的方法来使用 V ：

```makefile
ifeq ($(origin V), undefined)
MAKEFLAGS += --no-print-directory
endif
```

上面，介绍了 V OPTION，在 Makefile 中通过判断有没有定义 V ，来执行不同的 操作。其实还有一种 OPTION，这种 OPTION 的值在 Makefile 中是直接使用的，例如BINS。针对这种 OPTION，可以通过以下方式来使用：

```makefile
$ cat golang.mk

BINS ?= $(foreach cmd,${COMMANDS},$(notdir ${cmd}))
...
go.build: go.build.verify $(addprefix go.build., $(addprefix $(PLATFORM)., $(BINS)))
```

也就是说，通过 ?= 来判断 BINS 变量有没有被赋值，如果没有，则赋予等号后的值。接下来，就可以在 Makefile 规则中使用它。 

#### 技巧 10：定义环境变量 

可以在 Makefile 中定义一些环境变量，例如：

```makefile
$ cat golang.mk

GO := go
GO_SUPPORTED_VERSIONS ?= 1.13|1.14|1.15|1.16|1.17
GO_LDFLAGS += -X $(VERSION_PACKAGE).GitVersion=$(VERSION) \
	-X $(VERSION_PACKAGE).GitCommit=$(GIT_COMMIT) \
	-X $(VERSION_PACKAGE).GitTreeState=$(GIT_TREE_STATE) \
	-X $(VERSION_PACKAGE).BuildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
ifneq ($(DLV),)
	GO_BUILD_FLAGS += -gcflags "all=-N -l"
	LDFLAGS = ""
endif
GO_BUILD_FLAGS += -tags=jsoniter -ldflags "$(GO_LDFLAGS)"
...

$ cat common.mk
...
FIND := find . ! -path './third_party/*' ! -path './vendor/*'
XARGS := xargs --no-run-if-empty
```

这些环境变量和编程中使用宏定义的作用是一样的：只要修改一处，就可以使很多地方同时生效，避免了重复的工作。 

通常，可以将 GO、GO_BUILD_FLAGS、FIND 这类变量定义为环境变量。 

#### 技巧 11：自己调用自己 

在编写 Makefile 的过程中，可能会遇到这样一种情况：A-Target 目标命令中，需要完成操作 B-Action，而操作 B-Action 已经通过伪目标 B-Target 实现过。为了达到最大的代码复用度，这时候最好的方式是在 A-Target 的命令中执行 B-Target。方法如下：

```makefile
tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi
```

这里，通过 `$(MAKE) `调用了伪目标` tools.install.$* `。要注意的是，默认情况 下，Makefile 在切换目录时会输出以下信息：

```sh
$ make tools.install.codegen
===========> Installing codegen
make[1]: Entering directory 
`/home/colin/workspace/golang/src/github.com/marmotedu/iam'
make[1]: Leaving directory `/home/colin/workspace/golang/src/github.com/marmotedu/iam
```

如果觉得 Entering directory 这类信息很烦人，可以通过设置 MAKEFLAGS += --noprint-directory 来禁止 Makefile 打印这些信息。 

### 总结 

如果想要高效管理项目，使用 Makefile 来管理是目前的最佳实践。可以通过下面的 几个方法，来编写一个高质量的 Makefile。 

- 首先，需要熟练掌握 Makefile 的语法。建议重点掌握以下语法：Makefile 规则语法、伪目标、变量赋值、特殊变量、自动化变量。 
- 接着，需要提前规划 Makefile 要实现的功能。一个大型 Go 项目通常需要实现以下功能：代码生成类命令、格式化类命令、静态代码检查、 测试类命令、构建类命令、Docker 镜像打包类命令、部署类命令、清理类命令，等等。 
- 然后，还需要通过 Makefile 功能分类、文件分层、复杂命令脚本化等方式，来设计一个合理的 Makefile 结构。 
- 最后，还需要掌握一些 Makefile 编写技巧，例如：
  - 善用通配符、自动变量和函数；
  - 编写可扩展的 Makefile；
  - 使用带层级的命名方式，等等。
  - 通过这些技巧，可以进一步保证编写出一个高质量的 Makefile。 

### 课后练习

- 走读 IAM 项目的 Makefile 实现，看下 IAM 项目是如何通过 make tools.install 一键安装所有功能，通过 make tools.install.xxx 来指定安装 xxx 工具的。 
- 编写 Makefile 的时候，还用到过哪些编写技巧呢？



## Go 项目之 Makefile核心语法

熟练掌握 Makefile 语法的重要性，推荐学习陈皓老师编写的《跟我一起写 Makefile》 (PDF 重制版)。看到那么多 Makefile 语法，是不是有点被“劝退”的感觉？ 

虽然 Makefile 有很多语法，但不是所有的语法都需要熟练掌握，有些语法在 Go 项目中是很少用到的。要编写一个高质量的 Makefile，首先应该掌握一些核心的、最常用的语法知识。

这一讲就来具体介绍下 Go 项目中常用的 Makefile 语法和规则，快速打好最重要的基础。

Makefile 文件由三个部分组成，分别是 Makefile 规则、Makefile 语法和 Makefile 命令 （这些命令可以是 Linux 命令，也可以是可执行的脚本文件）。在这一讲里，会介绍下 Makefile 规则和 Makefile 语法里的一些核心语法知识。先来看下如何使用 Makefile 脚本。

### Makefile 的使用方法 

在实际使用过程中，一般是先编写一个 Makefile 文件，指定整个项目的编译规则，然后通过 Linux make 命令来解析该 Makefile 文件，实现项目编译、管理的自动化。

默认情况下，make 命令会在当前目录下，按照 GNUmakefile、makefile、Makefile 文件的顺序查找 Makefile 文件，一旦找到，就开始读取这个文件并执行。

大多数的 make 都支持“makefile”和“Makefile”这两种文件名，但建议使 用“Makefile”。因为这个文件名第一个字符大写，会很明显，容易辨别。

make 也支持 -f 和 --file 参数来指定其他文件名，比如 make -f golang.mk 或者 make --file golang.mk 。

### Makefile 规则介绍 

学习 Makefile，最核心的就是学习 Makefile 的规则。

规则是 Makefile 中的重要概念，它一般由目标、依赖和命令组成，用来指定源文件编译的先后顺序。Makefile 之所以受欢迎，核心原因就是 Makefile 规则，因为 Makefile 规则可以自动判断是否需要重新编译某个目标，从而确保目标仅在需要时编译。 

这一讲主要来看 Makefile 规则里的规则语法、伪目标和 order-only 依赖。 

#### 规则语法 

Makefile 的规则语法，主要包括 target、prerequisites 和 command，示例如下：

```makefile
target ...: prerequisites ...
  command
  ...
  ...
```

target，可以是一个 object file（目标文件），也可以是一个执行文件，还可以是一个标 签（label）。target 可使用通配符，当有多个目标时，目标之间用空格分隔。 

prerequisites，代表生成该 target 所需要的依赖项。当有多个依赖项时，依赖项之间用空格分隔。 

command，代表该 target 要执行的命令（可以是任意的 shell 命令）。

- 在执行 command 之前，默认会先打印出该命令，然后再输出命令的结果；如果不想打印出命令，可在各个 command 前加上@。 
- command 可以为多条，也可以分行写，但每行都要以 tab 键开始。另外，如果后一条命令依赖前一条命令，则这两条命令需要写在同一行，并用分号进行分隔。 
- 如果要忽略命令的出错，需要在各个 command 之前加上减号-。

只要 targets 不存在，或 prerequisites 中有一个以上的文件比 targets 文件新，那么 command 所定义的命令就会被执行，从而产生需要的文件，或执行期望的操 作。 

直接通过一个例子来理解下 Makefile 的规则吧。 

第一步，先编写一个 hello.c 文件。

```c
#include <stdio.h>
int main()
{
    printf("Hello World!\n");
    return 0;
}
```

第二步，在当前目录下，编写 Makefile 文件。

```makefile
hello: hello.o
	gcc -o hello hello.o

hello.o: hello.c
	gcc -c hello.c

clean:
	rm hello.o
```

第三步，执行 make，产生可执行文件。

```sh
$ make
gcc -c hello.c
gcc -o hello hello.o

$ ls
hello hello.c hello.o Makefile
```

上面的示例 Makefile 文件有两个 target，分别是 hello 和 hello.o，每个 target 都指定了构建 command。当执行 make 命令时，发现 hello、hello.o 文件不存在，就会执行 command 命令生成 target。 

第四步，不更新任何文件，再次执行 make。

```sh
$ make
make: 'hello' is up to date.
```

当 target 存在，并且 prerequisites 都不比 target 新时，不会执行对应的 command。 

第五步，更新 hello.c，并再次执行 make。

```sh
$ touch hello.c

$ make
gcc -c hello.c
gcc -o hello hello.o
```

当 target 存在，但 prerequisites 比 target 新时，会重新执行对应的 command。

第六步，清理编译中间文件。 

Makefile 一般都会有一个 clean 伪目标，用来清理编译中间产物，或者对源码目录做一些定制化的清理：

```sh
$ make clean
rm hello.o
```

可以在规则中使用通配符，make 支持三个通配符：*，? 和~，例如：

```sh
objects = *.o
print: *.c
	rm *.c
```

#### 伪目标 

接下来介绍下 Makefile 中的伪目标。Makefile 的管理能力基本上都是通过伪目标来实现的。 

在上面的 Makefile 示例中，定义了一个 clean 目标，这其实是一个伪目标，也就是说不会为该目标生成任何文件。因为伪目标不是文件，make 无法生成它的依赖关系， 也无法决定是否要执行它。 

通常情况下，需要显式地标识这个目标为伪目标。在 Makefile 中可以使用.PHONY来标识一个目标为伪目标：

```makefile
.PHONY: clean
clean:
	rm hello.o
```

伪目标可以有依赖文件，也可以作为“默认目标”，例如：

```makefile
.PHONY: all
all: lint test build
```

因为伪目标总是会被执行，所以其依赖总是会被决议。通过这种方式，可以达到同时执行所有依赖项的目的。 

#### order-only 依赖 

在上面介绍的规则中，只要 prerequisites 中有任何文件发生改变，就会重新构造 target。但是有时候，希望只有当 prerequisites 中的部分文件改变时，才重新构造 target。这时，可以通过 order-only prerequisites 实现。 

order-only prerequisites 的形式如下：

```makefile
targets : normal-prerequisites | order-only-prerequisites
  command
  ...
  ...
```

在上面的规则中，只有第一次构造 targets 时，才会使用 order-only-prerequisites。后面即使 order-only-prerequisites 发生改变，也不会重新构造 targets。 

只有 normal-prerequisites 中的文件发生改变时，才会重新构造 targets。这里，符号“ | ”后面的 prerequisites 就是 order-only-prerequisites。 

到这里，就介绍了 Makefile 的规则。接下来，再来看下 Makefile 中的一些核心语法知识。

### Makefile 语法概览 

因为 Makefile 的语法比较多，只介绍 Makefile 的核心语法，以及 IAM 项目的 Makefile 用到的语法，包括命令、变量、条件语句和函数。因为 Makefile 没有太多复杂的语法，掌握了这些知识点之后，再在实践中多加运用，融会贯通，就可以写出非常复杂、功能强大的 Makefile 文件了。

#### 命令 

Makefile 支持 Linux 命令，调用方式跟在 Linux 系统下调用命令的方式基本一致。默认情况下，make 会把正在执行的命令输出到当前屏幕上。但可以通过在命令前加@符号的方式，禁止 make 输出当前正在执行的命令。 

##### @ 符号

看一个例子。现在有这么一个 Makefile：

```makefile
.PHONY: test
test:
  echo "hello world"
```

执行 make 命令：

```sh
$ make test
echo "hello world"
hello world
```

可以看到，make 输出了执行的命令。很多时候，不需要这样的提示，因为更想看的是命令产生的日志，而不是执行的命令。这时就可以在命令行前加@，禁止 make 输出所执行的命令：

```makefile
.PHONY: test
test:
  @echo "hello world"
```

再次执行 make 命令：

```sh
$ make test
hello world
```

可以看到，make 只是执行了命令，而没有打印命令本身。这样 make 输出就清晰了很多。 

这里，建议在命令前都加@ 符号，禁止打印命令本身，以保证 Makefile 输出易于阅读的、有用的信息。 

##### - 符号

默认情况下，每条命令执行完 make 就会检查其返回码。如果返回成功（返回码为 0）， make 就执行下一条指令；如果返回失败（返回码非 0），make 就会终止当前命令。

很多时候，命令出错（比如删除了一个不存在的文件）时，并不想终止，这时就可以在命令行前加 - 符号，来让 make 忽略命令的出错，以继续执行下一条命令，比如：

```makefile
clean:
	-rm undefined.go
	
$ make clean
rm undefined.go
rm: undefined.go: No such file or directory
make: [clean] Error 1 (ignored)
```

#### 变量 

变量，可能是 Makefile 中使用最频繁的语法了，Makefile 支持变量赋值、多行变量和环境变量。另外，Makefile 还内置了一些特殊变量和自动化变量。 

##### 变量声明与引用

先来看下最基本的变量赋值功能。 Makefile 也可以像其他语言一样支持变量。在使用变量时，会像 shell 变量一样原地展开，然后再执行替换后的内容。 

Makefile 可以通过变量声明来声明一个变量，变量在声明时需要赋予一个初值，比如 ROOT_PACKAGE=github.com/marmotedu/iam。 

引用变量时可以通过`$()`或者`${}`方式引用。建议用`$()`方式引用变量，例如 $(ROOT_PACKAGE)，也建议整个 makefile 的变量引用方式保持一致。 

变量会像 bash 变量一样，在使用它的地方展开。比如：

```makefile
GO=go
build:
	$(GO) build -v .
```

展开后为：

```makefile
GO=go
build:
	go build -v .
```

接下来，介绍下 Makefile 中的 4 种变量赋值方法。

##### = 变量赋值

1. = 最基本的赋值方法。

例如：

```makefile
BASE_IMAGE = alpine:3.10
```

使用 = 进行赋值时，要注意下面这样的情况：

```makefile
A = a
B = $(A) b
A = c
```

B 最后的值为 c b，而不是 a b。也就是说，在用变量给变量赋值时，右边变量的取值，**取的是最终的变量值**。

##### := 变量赋值

2. :=直接赋值，赋予当前位置的值。

例如：

```makefile
A = a
B := $(A) b
A = c
```

B 最后的值为 a b。通过 := 的赋值方式，可以避免 = 赋值带来的潜在的不一致。

##### ?= 变量赋值

3. ?= 表示如果该变量没有被赋值，则赋予等号后的值。

例如：

```makefile
PLATFORMS ?= linux_amd64 linux_arm64
```

##### += 变量赋值

4. +=表示将等号后面的值添加到前面的变量上。

例如：

```makefile
MAKEFLAGS += --no-print-directory
```

##### define 多行变量

Makefile 还支持多行变量。可以通过 define 关键字设置多行变量，变量中允许换行。定义方式为：

```makefile
define 变量名
变量内容
...
endef
```

变量的内容可以包含函数、命令、文字或是其他变量。例如，可以定义一个 USAGE_OPTIONS 变量：

```makefile
define USAGE_OPTIONS

Options:
  ...
  BINS         The binaries to build. Default is all of cmd.
               ...
  ...
  V            Set to 1 enable verbose build. Default is 0.
endef
```

##### 环境变量

Makefile 还支持环境变量。在 Makefile 中，有两种环境变量，分别是 Makefile 预定义的环境变量和自定义的环境变量。 

其中，自定义的环境变量可以覆盖 Makefile 预定义的环境变量。默认情况下，Makefile 中定义的环境变量只在当前 Makefile 有效，如果想向下层传递（Makefile 中调用另一个 Makefile），需要使用 export 关键字来声明。 

下面的例子声明了一个环境变量，并可以在下层 Makefile 中使用：

```makefile
...
export USAGE_OPTIONS
...
```

##### 内置变量之特殊变量

此外，Makefile 还支持两种内置的变量：特殊变量和自动化变量。 

特殊变量是 make 提前定义好的，可以在 makefile 中直接引用。特殊变量列表如下：

![image-20211117000743656](IAM-document.assets/image-20211117000743656.png)

##### 内置变量之自动化变量

Makefile 还支持自动化变量。自动化变量可以提高编写 Makefile 的效率和质量。 

在 Makefile 的模式规则中，目标和依赖文件都是一系列的文件，那么如何书写一个命令，来完成从不同的依赖文件生成相对应的目标呢？ 

这时就可以用到自动化变量。所谓自动化变量，就是这种变量会把模式中所定义的一系列的文件自动地挨个取出，一直到所有符合模式的文件都取完为止。这种自动化变量只应出现在规则的命令中。Makefile 中支持的自动化变量见下表。

![image-20211117000958972](IAM-document.assets/image-20211117000958972.png)

上面这些自动化变量中，`$*`是用得最多的。`$* `对于构造有关联的文件名是比较有效的。如果目标中没有模式的定义，那么 `$* `也就不能被推导出。

但是，如果目标文件的后缀是 make 所识别的，那么 `$*` 就是除了后缀的那一部分。例如：如果目标是 foo.c ，因为.c 是 make 所能识别的后缀名，所以 `$*` 的值就是 foo。 

#### 条件语句 

Makefile 也支持条件语句。这里先看一个示例。

下面的例子判断变量ROOT_PACKAGE是否为空，如果为空，则输出错误信息，不为空则打印变量值：

```makefile
ifeq ($(ROOT_PACKAGE),)
$(error the variable ROOT_PACKAGE must be set prior to including golang.mk)
else
$(info the value of ROOT_PACKAGE is $(ROOT_PACKAGE))
endif
```

条件语句的语法为：

```makefile
# if ...
<conditional-directive>
<text-if-true>
endif

# if ... else ...
<conditional-directive>
<text-if-true>
else
<text-if-false>
endif
```

例如，判断两个值是否相等：

```makefile
ifeq 条件表达式
...
else
...
endif
```

- ifeq 表示条件语句的开始，并指定一个条件表达式。表达式包含两个参数，参数之间用逗号分隔，并且表达式用圆括号括起来。 
- else 表示条件表达式为假的情况。 endif 表示一个条件语句的结束，任何一个条件表达式都应该以 endif 结束。
- 表示条件关键字，有 4 个关键字：ifeq、ifneq、ifdef、ifndef。

为了加深理解，分别来看下这 4 个关键字的例子。

##### ifeq

1. ifeq：条件判断，判断是否相等。

例如：

```makefile
ifeq (<arg1>, <arg2>)
ifeq '<arg1>' '<arg2>'
ifeq "<arg1>" "<arg2>"
ifeq "<arg1>" '<arg2>'
ifeq '<arg1>' "<arg2>"
```

比较 arg1 和 arg2 的值是否相同，如果相同则为真。也可以用 make 函数 / 变量替代 arg1 或 arg2，例如 `ifeq ($(origin ROOT_DIR),undefined)` 或 `ifeq ($(ROOT_PACKAGE),) `。origin 函数会在之后专门讲函数的一讲中介绍到。

##### ifneq

2. ifneq：条件判断，判断是否不相等。

```makefile
ifneq (<arg1>, <arg2>)
ifneq '<arg1>' '<arg2>'
ifneq "<arg1>" "<arg2>"
ifneq "<arg1>" '<arg2>'
ifneq '<arg1>' "<arg2>"
```

比较 arg1 和 arg2 的值是否不同，如果不同则为真。

##### ifdef

3. ifdef：条件判断，判断变量是否已定义。

```makefile
ifdef <variable-name>
```

如果值非空，则表达式为真，否则为假。也可以是函数的返回值。

##### ifndef

4. ifndef：条件判断，判断变量是否未定义。

```makefile
ifndef <variable-name>
```

如果值为空，则表达式为真，否则为假。也可以是函数的返回值。

#### 函数 

Makefile 同样也支持函数，函数语法包括定义语法和调用语法。 

##### 自定义函数

先来看下自定义函数。 make 解释器提供了一系列的函数供 Makefile 调用，这些函 数是 Makefile 的预定义函数。可以通过 define 关键字来自定义一个函数。自定义函数的语法为：

```makefile
define 函数名
函数体
endef
```

例如，下面这个自定义函数：

```makefile
define Foo
  @echo "my name is $(0)"
  @echo "param is $(1)"
endef
```

define 本质上是定义一个多行变量，可以在 call 的作用下当作函数来使用，在其他位置使用只能作为多行变量来使用，例如：

```make
var := $(call Foo)
new := $(Foo)
```

自定义函数是一种过程调用，没有任何的返回值。可以使用自定义函数来定义命令的集合，并应用在规则中。 

##### 预定义函数

再来看下预定义函数。 刚才提到，make 编译器也定义了很多函数，这些函数叫作预定义函数，调用语法和变量类似，语法为：

```makefile
$(<function> <arguments>)
```

或者

```makefile
${<function> <arguments>}
```

<function> 是函数名，<arguments>是函数参数，参数间用逗号分割。函数的参数也可以是变量。

来看一个例子：

```makefile
PLATFORM = linux_amd64
GOOS := $(word 1, $(subst _, ,$(PLATFORM)))
```

上面的例子用到了两个函数：word 和 subst。

- word 函数有两个参数，1 和 subst 函数的输出。
- subst 函数将 PLATFORM 变量值中的 _ 替换成空格（替换后的 PLATFORM 值为 linux amd64）。
- word 函数取 linux amd64 字符串中的第一个单词。所以最后 GOOS 的 值为 linux。 

Makefile 预定义函数能够实现很多强大的功能，在编写 Makefile 的过程中，如果有功能需求，可以优先使用这些函数。如果想使用这些函数，那就需要知道有哪些函 数，以及它们实现的功能。

##### 常见的函数

常用的函数包括下面这些，需要先有个印象，以后用到时再来查看。

![常用的函数](IAM-document.assets/常用的函数.jpeg)

### 引入其他 Makefile 

除了 Makefile 规则、Makefile 语法之外，Makefile 还有很多特性，比如可以引入其他 Makefile、自动生成依赖关系、文件搜索等等。这里介绍一个 IAM 项目的 Makefile 用到的重点特性：引入其他 Makefile。 

Makefile 要结构化、层次化，这一点可以通过在项目根目录下的 Makefile 中引入其他 Makefile 来实现。 

在 Makefile 中，可以通过关键字 include，把别的 makefile 包含进来，类似于 C 语言的#include，被包含的文件会插入在当前的位置。include 用法为 `include <filename>`，示例如下：

```makefile
include scripts/make-rules/common.mk
include scripts/make-rules/golang.mk
```

include 也可以包含通配符include scripts/make-rules/*。make 命令会按下面的 顺序查找 makefile 文件：

- 如果是绝对或相对路径，就直接根据路径 include 进来。 
- 如果 make 执行时，有-I或--include-dir参数，那么 make 就会在这个参数所指定的目录下去找。
- 如果目录/include（一般是/usr/local/bin或/usr/include）存在的话，make 也会去找。

如果有文件没有找到，make 会生成一条警告信息，但不会马上出现致命错误，而是继续载入其他的文件。

一旦完成 makefile 的读取，make 会再重试这些没有找到或是不能读取的文件。

如果还是不行，make 才会出现一条致命错误信息。如果想让 make 忽略那些无法读取的文件继续执行，可以在 include 前加一个减号-，如-include 。 

### 总结 

在这一讲里，为了编写一个高质量的 Makefile，重点介绍了 Makefile 规则和 Makefile 语法里的一些核心语法知识。 

- 在讲 Makefile 规则时，主要学习了规则语法、伪目标和 order-only 依赖。掌握了这些 Makefile 规则，就掌握了 Makefile 中最核心的内容。 
- 在介绍 Makefile 的语法时，只介绍了 Makefile 的核心语法，以及 IAM 项目的 Makefile 用到的语法，包括命令、变量、条件语句和函数。

可能会觉得这些语法学习起来比较枯燥，但还是那句话，工欲善其事，必先利其器。希望熟练掌握 Makefile 的核心语法，为编写高质量的 Makefile 打好基础。





## Go 项目之研发流程管理

以研发流程为主线，来看下 IAM 项目是如何通过 Makefile 来高效管理项目的。不仅能更加深刻地理解如何设计研发流程，和如何基于 Makefile 高效地管理项目内容，还能得到很多可以直接用在实际操作中的经验、技巧。 

研发流程有很多阶段，其中的开发阶段和测试阶段是需要开发者深度参与的。所以在这一 讲中，会重点介绍这两个阶段中的 Makefile 项目管理功能，并且穿插一些 Makefile 的设计思路。

为了演示流程，这里先假设一个**场景**。有一个需求：给 IAM 客户端工具 iamctl 增加一个 helloworld 命令，该命令向终端打印 hello world。 

接下来，就来看下如何具体去执行研发流程中的每一步。首先，进入开发阶段。 

### 开发阶段 

开发阶段是开发者的主战场，完全由开发者来主导，它又可分为代码开发和代码提交两个子阶段。先来看下代码开发阶段。 

#### 代码开发 

拿到需求之后，首先需要开发代码。这时，就需要选择一个适合团队和项目的 Git 工 作流。因为 Git Flow 工作流比较适合大型的非开源项目，所以这里选择 Git Flow 工 作流。代码开发的具体步骤如下： 

##### 新建功能分支

第一步，基于 develop 分支，新建一个功能分支 feature/helloworld。

```sh
$ git checkout -b develop master
$ git checkout -b feature/helloworld develop

# 删除分支
$ git branch -d feature/helloworld
```

这里需要注意：新建的 branch 名要符合 Git Flow 工作流中的分支命名规则。否则，在 git commit 阶段，会因为 branch 不规范导致 commit 失败。IAM 项目的分支命令规则具体如下图所示：

![image-20211117202532409](IAM-document.assets/image-20211117202532409.png)

IAM 项目通过 pre-commit githooks 来确保分支名是符合规范的。在 IAM 项目根目录下执行 git commit 命令，git 会自动执行 pre-commit 脚本，该脚本会检查当前 branch 的名字是否符合规范。 

这里还有一个地方需要注意：git 不会提交 .git/hooks 目录下的 githooks 脚本，所 以需要通过以下手段，确保开发者 clone 仓库之后，仍然能安装指定的 githooks 脚本到 .git/hooks 目录：

```sh
# Copy githook scripts when execute makefile
COPY_GITHOOK:=$(shell cp -f githooks/* .git/hooks/)
```

上述代码放在 scripts/make-rules/common.mk 文件中，每次执行 make 命令时都会执行，可以确保 githooks 都安装到 .git/hooks 目录下。 

##### 添加命令模板

第二步，在 feature/helloworld 分支中，完成 helloworld 命令的添加。 

首先，通过 iamctl new helloworld 命令创建 helloworld 命令模板：

```sh
$ iamctl new helloworld -d internal/iamctl/cmd/helloworld
Command file generated: internal/iamctl/cmd/helloworld/helloworld.go
```

接着，编辑internal/iamctl/cmd/cmd.go文件，在源码文件中添加 helloworld.NewCmdHelloworld(f, ioStreams),，加载 helloworld 命令。这里将 helloworld 命令设置为Troubleshooting and Debugging Commands命令分组：

```go
import (
	"github.com/marmotedu/iam/internal/iamctl/cmd/helloworld"
)
	...
	{
		Message: "Troubleshooting and Debugging Commands:",
		Commands: []*cobra.Command{
			validate.NewCmdValidate(f, ioStreams),
			helloworld.NewCmdHelloworld(f, ioStreams),
		},
	}
```

这些操作中包含了 low code 的思想。要尽可能使用代码自动生成这一技术。这样做有两个好处：

- 一方面能够提高代码开发效率；
- 另一方面也能够保证规范，减少手动操作可能带来的错误。

所以这里，将 iamctl 的命令也模板化， 并通过 iamctl new 自动生成。

##### 生成代码

第三步，生成代码。

```sh
$ make gen
```

如果改动不涉及代码生成，可以不执行make gen操作。 make gen 执行的其实是 gen.run 伪目标：

```makefile
# cat ./scripts/make-rules/gen.mk
gen.run: gen.clean gen.errcode gen.docgo
```

可以看到，当执行 make gen.run 时，其实会先清理之前生成的文件，再分别自动生成 error code 和 doc.go 文件。

这里需要注意，通过make gen 生成的存量代码要具有幂等性。只有这样，才能确保每次生成的代码是一样的，避免不一致带来的问题。 

可以将更多的与自动生成代码相关的功能放在 gen.mk Makefile 中。例如：

- gen.docgo.doc，代表自动生成 doc.go 文件。 
- gen.ca.%，代表自动生成 iamctl、iam-apiserver、iam-authz-server 证书文件。

##### 版权检查

第四步，版权检查。

如果有新文件添加，还需要执行 make verify-copyright ，来检查新文件有没有添加版权头信息。

```sh
$ make verify-copyright
```

如果版权检查失败，可以执行 `make add-copyright` 自动添加版权头。添加版权信息只针对开源软件，如果软件不需要添加，就可以略过这一步。 

这里还有个 Makefile 编写技巧：如果 Makefile 的 command 需要某个命令，就可以使该目标依赖类似 tools.verify.addlicense 这种目标，tools.verify.addlicense 会检查该工具是否已安装，如果没有就先安装。

```makefile
.PHONY: copyright.verify
copyright.verify: tools.verify.addlicense
	...
tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi
	
.PHONY: install.addlicense
install.addlicense:
	@$(GO) get -u github.com/marmotedu/addlicense
```

通过这种方式，可以使 make copyright.verify 尽可能自动化，减少手动介入的概 率。 

##### 代码格式化

第五步，代码格式化。

```sh
$ make format
```

执行 make format 会依次执行以下格式化操作：

- 调用 gofmt 格式化代码。 
- 调用 goimports 工具，自动增删依赖的包，并将依赖包按字母序排序并分类。
- 调用 golines 工具，把超过 120 行的代码按 golines 规则，格式化成 <120 行的代码。
- 调用 go mod edit -fmt 格式化 go.mod 文件。

##### 静态代码检查

第六步，静态代码检查。

```sh
$ make lint
```

关于静态代码检查，在这里可以先了解代码开发阶段有这个步骤，至于如何操作，在下一讲详细介绍。 

##### 单元测试

第七步，单元测试。

```sh
$ make test
```

这里要注意，并不是所有的包都需要执行单元测试。可以通过如下命令，排除掉不需要单元测试的包：

```sh
go test `go list ./...|egrep -v $(subst $(SPACE),'|',$(sort $(EXCLUDE_TESTS)))`
```

运行该命令的目的，是把 `mock_.* .go `文件中的函数单元测试信息从 coverage.out 中删除。`mock_.*.go` 文件中的函数是不需要单元测试的，如果不删除，就会影响后面的单元测试覆盖率的计算。 

如果想检查单元测试覆盖率，请执行：

```sh
$ make cover
```

默认测试覆盖率至少为 60%，也可以在命令行指定覆盖率阈值为其他值，例如：

```sh
$ make cover COVERAGE=90
```

如果测试覆盖率不满足要求，就会返回以下错误信息：

```sh
test coverage is 62.1%
test coverage does not meet expectations: 90%, please add test cases!
make[1]: *** [go.test.cover] Error 1
make: *** [cover] Error 2
```

这里 make 命令的退出码为1。 

如果单元测试覆盖率达不到设置的阈值，就需要补充测试用例，否则禁止合并到 develop 和 master 分支。IAM 项目配置了 GitHub Actions CI 自动化流水线，CI 流水线会自动运行，检查单元测试覆盖率是否达到要求。 

##### 构建

第八步，构建。 

最后，执行make build命令，构建出cmd/目录下所有的二进制安装文件。

```sh
$ make build
```

make build 会自动构建 cmd/ 目录下的所有组件，如果只想构建其中的一个或多个组件，可以传入 BINS选项，组件之间用空格隔开，并用双引号引起来：

```sh
$ make build BINS="iam-apiserver iamctl"
```

到这里，就完成了代码开发阶段的全部操作。 

如果觉得手动执行的 make 命令比较多，可以直接执行 make 命令：

```sh
$ make
===========> Generating iam error code go source files
===========> Generating error code markdown documentation
===========> Generating missing doc.go for go packages
===========> Verifying the boilerplate headers for all files
===========> Formating codes
===========> Run golangci to lint source codes
===========> Run unit test
...
===========> Building binary iam-pump v0.7.2-24-g5814e7b for linux amd64
===========> Building binary iamctl v0.7.2-24-g5814e7b for linux amd64
...
```

直接执行make会执行伪目标 all 所依赖的伪目标 all: gen add-copyright format lint cover build，也即执行以下操作：生成代码、自动添加版权头、代码格式化、静态代码检查、单元测试、构建。 

这里需要注意一点：all 中依赖 cover，cover 实际执行的是 go.test.cover ，而 go.test.cover 又依赖 go.test ，所以 cover 实际上是先执行单元测试，再检查单元 测试覆盖率是否满足预设的阈值。 

最后补充一点，在开发阶段可以根据需要随时执行 make gen 、 make format 、 make lint 、 make cover 等操作，为的是能够提前发现问题并改正。

#### 代码提交 

代码开发完成之后，就需要将代码提交到远程仓库，整个流程分为以下几个步骤。

##### 提交 feature 分支，push 到远端仓库 

第一步，开发完后，将代码提交到 feature/helloworld 分支，并 push 到远端仓库。

```sh
$ git add internal/iamctl/cmd/helloworld internal/iamctl/cmd/cmd.go
$ git commit -m "feat: add new iamctl command 'helloworld'"
$ git push origin feature/helloworld

# 如果报错，.git/hooks/commit-msg: line 12: go-gitlint: command not found
# 执行： go get github.com/llorllale/go-gitlint/cmd/go-gitlint
```

这里建议只添加跟feature/helloworld相关的改动，这样就知道一个 commit 做 了哪些变更，方便以后追溯。所以，不建议直接执行git add . 这类方式提交改动。 

在提交 commit 时，commit-msg githooks 会检查 commit message 是否符合 Angular Commit Message 规范，如果不符合会报错。commit-msage 调用了go-gitlint 来检查 commit message。go-gitlint 会读取 .gitlint 中配置的 commit message 格式：

```sh
--subject-regex=^((Merge branch.*of.*)|((revert: )?(feat|fix|perf|style|refactor|test|ci|docs|chore)(\(.+\))?: [^A-Z].*[^.]$))
--subject-maxlen=72
--body-regex=^([^\r\n]{0,72}(\r?\n|$))*$
```

IAM 项目配置了 GitHub Actions，当有代码被 push 后，会触发 CI 流水线，流水线会执行 make all目标。GitHub Actions CI 流程执行记录如下图：

![image-20211117214858491](IAM-document.assets/image-20211117214858491.png)

如果 CI 不通过，就需要修改代码，直到 CI 流水线通过为止。 这里，来看下 GitHub Actions 的配置：

```sh
# .github/workflows/iamci.yaml

name: IamCI

on:
  push:
    branchs:
    - '*'
  pull_request:
  	types: [opened, reopened]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
    	uses: actions/setup-go@v2
      with:
      	go-version: 1.16
      	
    - name: all
    	run: make
```

可以看到，GitHub Actions 实际上执行了 3 步：拉取代码、设置 Go 编译环境、执行 make 命令（也就是执行 make all 目标）。

 GitHub Actions 也执行了 make all 目标，和手动操作执行的 make all 目标保持一 致，这样做是为了让线上的 CI 流程和本地的 CI 流程完全保持一致。这样，在本地 执行 make 命令通过后，在线上也会通过。保持一个一致的执行流程和执行结果很重要。 否则，本地执行 make 通过，但是线上却不通过，岂不很让人头疼？ 

##### 提交 pull request

第二步，提交 pull request。 

登陆 GitHub，基于 feature/helloworld 创建 pull request，并指定 Reviewers 进行 code review。具体操作如下图：

![image-20211117215107770](IAM-document.assets/image-20211117215107770.png)

当有新的 pull request 被创建后，也会触发 CI 流水线。 

##### 通知 reviewers

第三步，创建完 pull request 后，就可以通知 reviewers 来 review 代码，GitHub 也会发站内信。 

##### 代码 review 

第四步，Reviewers 对代码进行 review。 

Reviewer 通过 review github diff 后的内容，并结合 CI 流程是否通过添加评论，并选择 Comment（仅评论）、Approve（通过）、Request Changes（不通过，需要修改）， 如下图所示：

![image-20211117215705447](IAM-document.assets/image-20211117215705447.png)

如果 review 不通过，feature 开发者可以直接在 feature/helloworld 分支修正代码，并 push 到远端的 feature/helloworld 分支，然后通知 reviewers 再次 review。因为有 push 事件发生，所以会触发 GitHub Actions CI 流水线。 

##### 合并 develop 分支

第五步，code review 通过后，maintainer 就可以将新的代码合并到 develop 分支。 使用 Create a merge commit 的方式，将 pull request 合并到 develop 分支，如下图所示：

![image-20211117220008089](IAM-document.assets/image-20211117220008089.png)

Create a merge commit 的实际操作是 git merge --no-ff，feature/helloworld 分 支上所有的 commit 都会加到 develop 分支上，并且会生成一个 merge commit。使用这种方式，可以清晰地知道是谁做了哪些提交，回溯历史的时候也会更加方便。 

##### 触发 CI 流程

第六步，合并到 develop 分支后，触发 CI 流程。 

到这里，开发阶段的操作就全部完成了，整体流程如下：

![image-20211117220414938](IAM-document.assets/image-20211117220414938.png)

合并到 develop 分支之后，就可以进入开发阶段的下一阶段，也就是测试阶段了。

### 测试阶段 

在测试阶段，开发人员主要负责提供测试包和修复测试期间发现的 bug，这个过程中也可能会发现一些新的需求或变动点，所以需要合理评估这些新的需求或变动点是否要放在当前迭代修改。 

测试阶段的操作流程如下。 

#### 创建 release 分支

第一步，基于 develop 分支，创建 release 分支，测试代码。

```sh
$ git checkout -b release/1.0.0 develop
$ make
```

#### 提交测试

第二步，提交测试。 

将 release/1.0.0 分支的代码提交给测试同学进行测试。这里假设一个测试失败的场景：要求打印“hello world”，但打印的是“Hello World”，需要修复。那具体应该怎么操作呢？

可以直接在 release/1.0.0 分支修改代码，修改完成后，本地构建并提交代码：

```sh
$ make
$ git add internal/iamctl/cmd/helloworld/helloworld.go
$ git commit -m "fix: fix helloworld print bug"
$ git push origin release/1.0.0
```

push 到 release/1.0.0 后，GitHub Actions 会执行 CI 流水线。如果流水线执行成功，就将代码提供给测试；如果测试不成功，再重新修改，直到流水线执行成功。 

测试同学会对 release/1.0.0 分支的代码进行充分的测试，例如功能测试、性能测试、集成测试、系统测试等。 

#### 合并 master 分支

第三步，测试通过后，将功能分支合并到 master 分支和 develop 分支。

```sh
$ git checkout develop
$ git merge --no-ff release/1.0.0
$ git checkout master
$ git merge --no-ff release/1.0.0
$ git tag -a v1.0.0 -m "add print hello world" # master分支打tag
```

到这里，测试阶段的操作就基本完成了。测试阶段的产物是 master/develop 分支的代码。 

#### 删除 feature 分支

第四步，删除 feature/helloworld 分支，也可以选择性删除 release/1.0.0 分支。

代码都合并入 master/develop 分支后，feature 开发者可以选择是否要保留 feature。不过，如果没有特别的原因，建议删掉，因为 feature 分支太多的话，不仅看起来很乱，还会影响性能，删除操作如下：

```sh
$ git branch -d feature/helloworld
```

### IAM 项目的 Makefile 项目管理技巧 

在上面的内容中，以研发流程为主线，亲身体验了 IAM 项目的 Makefile 项目管理功 能。这些是最应该掌握的核心功能，但 IAM 项目的 Makefile 还有很多功能和设计技 巧。

接下来，会分享一些很有价值的 Makefile 项目管理技巧。 

#### help 自动解析 

因为随着项目的扩展，Makefile 大概率会不断加入新的管理功能，这些管理功能也需要加入到 make help 输出中。但如果每添加一个目标，都要修改 make help 命令，就比较麻烦，还容易出错。

所以这里，通过自动解析的方式，来生成make help输出：

```makefile
## help: Show this help info.
.PHONY: help
help: Makefile
  @echo -e "\nUsage: make <TARGETS> <OPTIONS> ...\n\nTargets:"
  @sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
  @echo "$$USAGE_OPTIONS"
```

目标 help 的命令中，通过 `sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /' `命令，自动解析 Makefile 中 `##` 开头的注释行，从而自动生成 make help 输出。 

#### Options 中指定变量值 

通过以下赋值方式，变量可以在 Makefile options 中被指定：

```sh
ifeq ($(origin COVERAGE),undefined)
COVERAGE := 60
endif
```

例如，如果执行make ，则 COVERAGE 设置为默认值 60；如果执行make COVERAGE=90 ，则 COVERAGE 值为 90。通过这种方式，可以更灵活地控制Makefile 的行为。 

#### 自动生成 CHANGELOG 

一个项目最好有 CHANGELOG 用来展示每个版本之间的变更内容，作为 Release Note 的一部分。但是，如果每次都要手动编写 CHANGELOG，会很麻烦，也不容易坚持，所以这里可以借助 git-chglog工具来自动生成。 

IAM 项目的 git-chglog 工具的配置文件放在 .chglog 目录下，在学习 git-chglog 工具时，可以参考下。

#### 自动生成版本号 

一个项目也需要有一个版本号，当前用得比较多的是语义化版本号规范。但如果靠开发者手动打版本号，工作效率低不说，经常还会出现漏打、打的版本号不规范等问题。

所以最好的办法是，版本号也通过工具自动生成。在 IAM 项目中，采用了 gsemver工具来自动生成版本号。 

整个 IAM 项目的版本号，都是通过 `scripts/ensure_tag.sh` 脚本来生成的：

```sh
version=v`gsemver bump`
if [ -z "`git tag -l $version`" ];then
	git tag -a -m "release version $version" $version
fi
```

在 scripts/ensure_tag.sh 脚本中，通过 gsemver bump 命令来自动化生成语义化的版本号，并执行 git tag -a 给仓库打上版本号标签，gsemver 命令会根据 Commit Message 自动生成版本号。 

之后，Makefile 和 Shell 脚本用到的所有版本号均统一使用`scripts/make-rules/common.mk` 文件中的 VERSION 变量：

```makefile
VERSION := $(shell git describe --tags --always --match='v*')
```

上述的 Shell 命令通过 git describe 来获取离当前提交最近的 tag（版本号）。 

在执行 git describe 时，如果符合条件的 tag 指向最新提交，则只显示 tag 的名字， 否则会有相关的后缀，来描述该 tag 之后有多少次提交，以及最新的提交 commit id。例如：

```SH
$ git describe --tags --always --match='v*'
v1.0.0-3-g1909e47
```

这里解释下版本号中各字符的含义：

- 3：表示自打 tag v1.0.0 以来有 3 次提交。 
- g1909e47：g 为 git 的缩写，在多种管理工具并存的环境中很有用处。 
- 1909e47：7 位字符表示为最新提交的 commit id 前 7 位。

最后解释下参数：

- --tags，不要只使用带注释的标签，而要使用refs/tags名称空间中的任何标签。 
- --always，显示唯一缩写的提交对象作为后备。 
- --match，只考虑与给定模式相匹配的标签。

#### 保持行为一致 

上面介绍了一些管理功能，例如检查 Commit Message 是否符合规范、自动生成 CHANGELOG、自动生成版本号。这些可以通过 Makefile 来操作，也可以手动执 行。

例如，通过以下命令，检查 IAM 的所有 Commit 是否符合 Angular Commit Message 规范：

```sh
$ go-gitlint
b62db1f: subject does not match regex [^(revert: )?(feat|fix|perf|style|refactor|test|ci|docs|chore)(\(.+\))?: [^A-Z].*[^.]$]
```

也可以通过以下命令，手动来生成 CHANGELOG：

```sh
$ git-chglog v1.0.0 CHANGELOG/CHANGELOG-1.0.0.md
```

还可以执行 gsemver 来生成版本号：

```sh
$ gsemver bump
1.0.1
```

这里要强调的是，要保证不管使用手动操作，还是通过 Makefile 操作，都要确保 git commit message 规范检查结果、生成的 CHANGELOG、生成的版本号是一致的。这需要采用同一种操作方式。 

### 总结 

在整个研发流程中，需要开发人员深度参与的阶段有两个，分别是开发阶段和测试阶段。 

- 在开发阶段，开发者完成代码开发之后，通常需要执行生成代码、版权检查、代码格式化、静态代码检查、单元测试、构建等操作。
- 可以将这些操作集成在 Makefile 中，来提高效率，并借此统一操作。 
- 另外，IAM 项目在编写 Makefile 时也采用了一些技巧，例如make help 命令中，help 信息是通过解析 Makefile 文件的注释来完成的；
- 可以通过 git-chglog 自动生成 CHANGELOG；
- 通过 gsemver 自动生成语义化的版本号等。 

### 课后练习

- 看下 IAM 项目的 make dependencies 是如何实现的，这样实现有什么好处？ 
- IAM 项目中使用 了gofmt 、goimports 、golines 3 种格式化工具，思考下，还有没有其他格式化工具值得集成在 make format 目标的命令中？



## Go 项目之静态代码检查

在讲代码开发的具体步骤时，提到了静态代码检查，就来详细讲讲如何执行静态代码检查。 

在做 Go 项目开发的过程中，肯定需要对 Go 代码做静态代码检查。虽然 Go 命令提 供了 go vet 和 go tool vet，但是它们检查的内容还不够全面，需要一种更加强大的 静态代码检查工具。 

其实，Go 生态中有很多这样的工具，也不乏一些比较优秀的。golangci-lint，是目前使用最多，也最受欢迎的静态代码检查工具，IAM 实战项目也用到了它。

接下来，就从 golangci-lint 的优点、golangci-lint 提供的命令和选项、golangci-lint 的配置这三个方面来介绍下它。

在了解这些基础知识后，会使用 golangci-lint 进行静态代码检查，熟悉操作，在这个基础上，再把使用 golangci-lint 时总结的一些经验技巧分享。 

### 为什么选择 golangci-lint 做静态代码检查？ 

选择 golangci-lint，是因为它具有其他静态代码检查工具不具备的一些优点。它的核心优点至少有这些：

- 速度非常快：golangci-lint 是基于 gometalinter 开发的，但是平均速度要比 gometalinter 快 5 倍。golangci-lint 速度快的原因有三个：可以并行检查代码；可以复用 go build 缓存；会缓存分析结果。 
- 可配置：支持 YAML 格式的配置文件，让检查更灵活，更可控。 
- IDE 集成：可以集成进多个主流的 IDE，例如 VS Code、GNU Emacs、Sublime Text、Goland 等。 
- linter 聚合器：1.41.1 版本的 golangci-lint 集成了 76 个 linter，不需要再单独安装这 76 个 linter。并且 golangci-lint 还支持自定义 linter。 
- 最小的误报数：golangci-lint 调整了所集成 linter 的默认设置，大幅度减少了误报。 
- 良好的输出：输出的结果带有颜色、代码行号和 linter 标识，易于查看和定位。

下图是一个 golangci-lint 的检查结果：

![image-20211118000919320](IAM-document.assets/image-20211118000919320.png)

可以看到，输出的检查结果中包括如下信息：

- 检查出问题的源码文件、行号和错误行内容。 
- 出问题的原因，也就是打印出不符合检查规则的原因。
- 报错的 linter。

通过查看 golangci-lint 的输出结果，可以准确地定位到报错的位置，快速弄明白报错的原因，方便开发者修复。 

除了上述优点之外， golangci-lint 还有一个非常大的优点：当前更新迭代速度很 快，不断有新的 linter 被集成到 golangci-lint 中。有这么全的 linter 为代码保驾护 航，在交付代码时肯定会更有自信。 

目前，有很多公司 / 项目使用了 golangci-lint 工具作为静态代码检查工具，例如 Google、Facebook、Istio、Red Hat OpenShift 等。 

### golangci-lint 提供了哪些命令和选项？ 

在使用之前，首先需要安装 golangci-lint。golangci-lint 的安装方法也很简单，只需要执行以下命令，就可以安装了。

```sh
$ go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1
$ golangci-lint version # 输出 golangci-lint 版本号，说明安装成功
golangci-lint has version v1.39.0 built from (unknown, mod sum: "h1:aAUjdBxARwkGLd5PU0vKuym281f2rFOyqh3GB4nXcq8="
```

这里注意，为了避免安装失败，强烈建议安装 golangci-lint releases page 中的指定版本，例如 v1.41.1。 

另外，还建议定期更新 golangci-lint 的版本，因为该项目正在被积极开发并不断改进。 

安装之后，就可以使用了。可以通过执行 `golangci-lint -h` 查看其用法， golangci-lint 支持的子命令见下表：

![image-20211118001511624](IAM-document.assets/image-20211118001511624.png)

此外，golangci-lint 还支持一些全局选项。全局选项是指适用于所有子命令的选项， golangci-lint 支持的全局选项如下：

![image-20211118001655735](IAM-document.assets/image-20211118001655735.png)

接下来，就详细介绍下 golangci-lint 支持的核心子命令：run、cache、completion、 config、linters。 

#### golangci-lint 命令

##### run 命令 

run 命令执行 golangci-lint，对代码进行检查，是 golangci-lint 最为核心的一个命令。 

run 没有子命令，但有很多选项。run 命令的具体使用方法，会在讲解如何执行静态代码检查的时候详细介绍。

##### cache 命令 

cache 命令用来进行缓存控制，并打印缓存的信息。它包含两个子命令：

- clean 用来清除 cache，当觉得 cache 的内容异常，或者 cache 占用空间过大时， 可以通过 `golangci-lint cache clean` 清除 cache。 

- status 用来打印 cache 的状态，比如 cache 的存放目录和 cache 的大小，例如：

  - ```sh
    $ golangci-lint cache status
    Dir: /home/going/.cache/golangci-lint
    Size: 1.7MiB
    ```

##### completion 命令 

completion 命令包含 4 个子命令 bash、fish、powershell 和 zsh，分别用来输出 bash、fish、powershell 和 zsh 的自动补全脚本。 

下面是一个配置 bash 自动补全的示例：

```sh
$ golangci-lint completion bash > ~/.golangci-lint.bash
$ echo "source '$HOME/.golangci-lint.bash'" >> ~/.bashrc
$ source ~/.bashrc

# 如果权限报错 -bash: /home/going/.bashrc: Permission denied，执行
$ sudo chmod 777 ~/.bashr
```

执行完上面的命令，键入如下命令，即可自动补全子命令：

```sh
$ golangci-lint comp<TAB>
```

上面的命令行会自动补全为 `golangci-lint completion` 。

##### config 命令 

config 命令可以打印 golangci-lint 当前使用的配置文件路径，例如：

```sh
$ golangci-lint config path
.golangci.yaml
```

##### linters 命令

linters 命令可以打印出 golangci-lint 所支持的 linter，并将这些 linter 分成两类，分别是配置为启用的 linter 和配置为禁用的 linter，例如：

```sh
$ golangci-lint linters
Enabled by your configuration linters:
...
deadcode: Finds unused code [fast: true, auto-fix: false]
...
Disabled by your configuration linters:
exportloopref: checks for pointers to enclosing loop variables [fast: true, auto-fix: false]
...
```

上面介绍了 golangci-lint 提供的命令，接下来，再来看下 golangci-lint 的配置。

#### golangci-lint 配置 

和其他 linter 相比，golangci-lint 一个非常大的优点是使用起来非常灵活，这要得益于它对自定义配置的支持。 

golangci-lint 支持两种配置方式，分别是命令行选项和配置文件。

- 如果 bool/string/int 的选项同时在命令行选项和配置文件中被指定，命令行的选项就会覆盖配置文件中的选项。
- 如果是 slice 类型的选项，则命令行和配置中的配置会进行合并。 

##### 命令行选项

golangci-lint run 支持很多命令行选项，可通过 `golangci-lint run -h ` 查看，这 里选择一些比较重要的选项进行介绍，见下表：

![image-20211118003525635](IAM-document.assets/image-20211118003525635.png)

![image-20211118003612088](IAM-document.assets/image-20211118003612088.png)

##### 配置文件

此外，还可以通过 golangci-lint 配置文件进行配置，默认的配置文件名 为 .golangci.yaml、.golangci.toml、.golangci.json，可以通过 -c 选项指定配置文件名。 通过配置文件，可以实现下面几类功能：

- golangci-lint 本身的一些选项，比如超时、并发，是否检查 *_test.go 文件等。
- 配置需要忽略的文件和文件夹。 
- 配置启用哪些 linter，禁用哪些 linter。 
- 配置输出格式。 
- golangci-lint 支持很多 linter，其中有些 linter 支持一些配置项，这些配置项可以在配置文件中配置。 
- 配置符合指定正则规则的文件可以忽略的 linter。 
- 设置错误严重级别，像日志一样，检查错误也是有严重级别的。

更详细的配置内容，可以参考 Configuration。另外，也可以参考 IAM 项目的 golangci-lint 配置 .golangci.yaml。.golangci.yaml 里面的一些配置，建议一定要设置，具体如下：

```yaml
run:
  skip-dirs: # 设置要忽略的目录
    - util
    - .*~
    - api/swagger/docs
  skip-files: # 设置不需要检查的go源码文件，支持正则匹配，这里建议包括：_test.go
    - ".*\\.my\\.go$"
    - _test.go

linters-settings:
  errcheck:
    check-type-assertions: false  # 这里建议设置为true，如果确实不需要检查，可以写成`num, _ := strconv.Atoi(numStr)`
    check-blank: false
  gci:
  	# 将以`github.com/marmotedu/iam`开头的包放在第三方包后面
    local-prefixes: github.com/marmotedu/iam
  godox:
    keywords: # 建议设置为BUG、FIXME、OPTIMIZE、HACK
      - BUG
      - FIXME
      - OPTIMIZE 
      - HACK 
  goimports:
  	# 设置哪些包放在第三方包后面，可以设置多个包，逗号隔开
    local-prefixes: github.com/marmotedu/iam
  gomoddirectives: # 设置允许在go.mod中replace的包
    replace-local: true
    replace-allow-list:
    	- github.com/coreos/etcd
      - google.golang.org/grpc
      - github.com/marmotedu/api
      - github.com/marmotedu/component-base
      - github.com/marmotedu/marmotedu-sdk-go
  gomodguard: # 下面是根据需要选择可以使用的包和版本，建议设置
    allowed:
      modules:  
        - gorm.io/gorm
        - gorm.io/driver/mysql
        - k8s.io/klog
      domains:                                                        # List of allowed module domains
        - google.golang.org
        - gopkg.in
        - golang.org
        - github.com
        - go.uber.org
        - go.etcd.io
    blocked:
      modules:
        - github.com/pkg/errors:
            recommendations:
              - github.com/marmotedu/errors
            reason: "`github.com/marmotedu/errors` is the log package used by marmotedu projects."
      versions:
        - github.com/MakeNowJust/heredoc:
            version: "> 2.0.9"
            reason: "use the latest version"
      local_replace_directives: false   
  lll:
    line-length: 240 # 这里可以设置为240，240一般是够用的
  importas: # 设置包的alias，根据需要设置
    # using `jwt` alias for `github.com/appleboy/gin-jwt/v2` package
    jwt: github.com/appleboy/gin-jwt/v2
    # using `metav1` alias for `github.com/marmotedu/component-base/pkg/meta/v1` package
    metav1: github.com/marmotedu/component-base/pkg/meta/v1
```

需要注意的是，golangci-lint 不建议使用 enable-all: true 选项，为了尽可能使用最 全的 linters，可以使用以下配置：

```yaml
linters:
  disable-all: true
  enable: # enable下列出 <期望的所有linters>
    - typecheck
    - ...
```

<期望的所有linters> =  - <不期望执行的 linters>，可以通过执行以下命令来获取：

```sh
$ ./scripts/print_enable_linters.sh
  - asciicheck
  - bodyclose
  - cyclop
  - deadcode
  - ...
```

将以上输出结果替换掉 .golangci.yaml  配置文件中的 linters.enable 部分即可。 

上面介绍了与 golangci-lint 相关的一些基础知识，接下来就详细展示下，如何使用 golangci-lint 进行静态代码检查。 

### 如何使用 golangci-lint 进行静态代码检查？ 

要对代码进行静态检查，只需要执行 golangci-lint run 命令即可。

#### golangci-lint 用法

接下来，会介绍 5 种常见的 golangci-lint 使用方法。

1. 对当前目录及子目录下的所有 Go 文件进行静态代码检查：

```sh
$ golangci-lint run
```

命令等效于golangci-lint run ./...。

2. 对指定的 Go 文件或者指定目录下的 Go 文件进行静态代码检查：

```sh
$ golangci-lint run dir1 dir2/... dir3/file1.go
```

这里需要注意：上述命令不会检查 dir1 下子目录的 Go 文件，如果想递归地检查一个目录，需要在目录后面追加/...，例如：dir2/...。

3. 根据指定配置文件，进行静态代码检查：

```sh
 $ golangci-lint run -c .golangci.yaml ./..
```

4. 运行指定的 linter：

golangci-lint 可以在不指定任何配置文件的情况下运行，这会运行默认启用的 linter，可以通过 `golangci-lint help linters` 查看它。 

可以传入参数 -E/--enable 来使某个 linter 可用，也可以使用 -D/--disable 参数来使某个 linter 不可用。下面的示例仅仅启用了 errcheck linter：

```sh
$ golangci-lint run --no-config --disable-all -E errcheck ./...
```

这里需要注意，默认情况下，golangci-lint 会从当前目录一层层往上寻找配置文件 名.golangci.yaml、.golangci.toml、.golangci.json 直到根（/）目录。如果找到，就以找到的配置文件作为本次运行的配置文件，所以为了防止读取到未知的配置文 件，可以用 --no-config 参数使 golangci-lint 不读取任何配置文件。

5. 禁止运行指定的 liner：

如果想禁用某些 linter，可以使用-D选项。

```sh
$ golangci-lint run --no-config -D godot,errcheck
```

在使用 golangci-lint 进行代码检查时，可能会有很多误报。所谓的误报，其实是希望 golangci-lint 的一些 linter 能够容忍某些 issue。那么如何尽可能减少误报呢？golangcilint 也提供了一些途径，建议使用下面这三种：

- 在命令行中添加-e参数，或者在配置文件的 issues.exclude 部分设置要排除的检查错误。也可以使用 issues.exclude-rules 来配置哪些文件忽略哪些 linter。
- 通过 run.skip-dirs、run.skip-files 或者 issues.exclude-rules配置项，来忽略指定目录下的所有 Go 文件，或者指定的 Go 文件。 
- 通过在 Go 源码文件中添加 //nolint 注释，来忽略指定的代码行。

因为 golangci-lint 设置了很多 linters，对于一个大型项目，启用所有的 linter 会检查出很多问题，并且每个项目对 linter 检查的粒度要求也不一样，所以 glangci-lint使用 nolint 标记来开关某些检查项，不同位置的 nolint 标记效果也会不一样。接下来介绍 nolint 的几种用法。

#### nolint 用法

1. 忽略某一行所有 linter 的检查

```sh
var bad_name int //nolint
```

2. 忽略某一行指定 linter 的检查，可以指定多个 linter，用逗号 , 隔开。

```sh
var bad_name int //nolint:golint,unused
```

3. 忽略某个代码块的检查。

```sh
// nolint
func allIssuesInThisFunctionAreExcluded() *string {
  // ...
}

// nolint:govet
var (
  a int
  b int
)
```

4. 忽略某个文件的指定 linter 检查。

在 package xx 上面一行添加 //nolint 注释。

```sh
//nolint:unparam
package pkg
...
```

在使用 nolint 的过程中，有 3 个地方需要注意。 

- 首先，如果启用了 nolintlint，就需要在//nolint后面添加 nolint 的原因// xxxx。 
- 其次，使用的应该是//nolint而不是// nolint。因为根据 Go 的规范，需要程序读 取的注释 // 后面不应该有空格。 
- 最后，如果要忽略所有 linter，可以用//nolint；如果要忽略某个指定的 linter，可以用//nolint:,。

### golangci-lint 使用技巧 

在使用 golangci-lint 时，总结了一些经验技巧，放在这里参考，希望能够更好地使用 golangci-lint。 

#### 技巧 1：第一次修改，可以按目录修改

如果第一次使用 golangci-lint 检查代码，一定会有很多错误。为了减轻修改的压 力，可以按目录检查代码并修改。这样可以有效减少失败条数，减轻修改压力。

 当然，如果错误太多，一时半会儿改不完，想以后慢慢修改或者干脆不修复存量的 issues，那么可以使用 golangci-lint 的 --new-from-rev 选项，只检查新增的 code，例如：

```sh
$ golangci-lint run --new-from-rev=HEAD~1
```

#### 技巧 2：按文件修改，减少文件切换次数，提高修改效率

如果有很多检查错误，涉及很多文件，建议先修改一个文件，这样就不用来回切换文件。 可以通过 grep 过滤出某个文件的检查失败项，例如：

```sh
$ golangci-lint run ./...|grep pkg/storage/redis_cluster.go
pkg/storage/redis_cluster.go:16:2: "github.com/go-redis/redis/v7" imported but
pkg/storage/redis_cluster.go:82:28: undeclared name: `redis` (typecheck)
pkg/storage/redis_cluster.go:86:14: undeclared name: `redis` (typecheck)
...
```

#### 技巧 3：把 linters-setting.lll.line-length 设置得大一些 

在 Go 项目开发中，为了易于阅读代码，通常会将变量名 / 函数 / 常量等命名得有意义， 这样很可能导致每行的代码长度过长，很容易超过 lll linter 设置的默认最大长度 80。

这里建议将linters-setting.lll.line-length设置为 120/240。 

#### 技巧 4：尽可能多地使用 golangci-lint 提供的 linter

 golangci-lint 集成了很多 linters，可以通过如下命令查看：

```sh
$ golangci-lint linters
Enabled by your configuration linters:
deadcode: Finds unused code [fast: true, auto-fix: false]
...
varcheck: Finds unused global variables and constants [fast: true, auto-fix: false]

Disabled by your configuration linters:
asciicheck: Simple linter to check that your code does not contain non-ASCII identifiers [fast: true, auto-fix: false]
...
wsl: Whitespace Linter - Forces you to use empty lines! [fast: true, auto-fix: false]
```

这些 linter 分为两类，一类是默认启用的，另一类是默认禁用的。每个 linter 都有两个属 性：

- fast：true/false，如果为 true，说明该 linter 可以缓存类型信息，支持快速检查。因为第一次缓存了这些信息，所以后续的运行会非常快。 
- auto-fix：true/false，如果为 true 说明该 linter 支持自动修复发现的错误；如果为 false 说明不支持自动修复。

如果配置了 golangci-lint 配置文件，则可以通过命令 golangci-lint help linters 查看在当前配置下启用和禁用了哪些 linter。golangci-lint 也支持自定义 linter 插件，具 体可以参考：New linters。 

在使用 golangci-lint 的时候，要尽可能多的使用 linter。使用的 linter 越多，说明检 查越严格，意味着代码越规范，质量越高。如果时间和精力允许，建议打开 golangci-lint 提供的所有 linter。 

#### 技巧 5：每次修改代码后，都要执行 golangci-lint 

每次修改完代码后都要执行 golangci-lint，一方面可以及时修改不规范的地方，另一方面 可以减少错误堆积，减轻后面的修改压力。

#### 技巧 6：建议在根目录下放一个通用的 golangci-lint 配置文件。

在/目录下存放通用的 golangci-lint 配置文件，可以不用为每一个项目都配置 golangci-lint。当需要为某个项目单独配置 golangci-lint 时，只需在该项目根目录下增加一个项目级别的 golangci-lint 配置文件即可。 

### 总结 

Go 项目开发中，对代码进行静态代码检查是必要的操作。当前有很多优秀的静态代码检查工具，但 golangci-lint 因为具有检查速度快、可配置、少误报、内置了大量 linter 等优点，成为了目前最受欢迎的静态代码检查工具。 

golangci-lint 功能非常强大，支持诸如 run、cache、completion、linters 等命令。其中最常用的是 run 命令，run 命令可以通过以下方式来进行静态代码检查：

```sh
$ golangci-lint run # 对当前目录及子目录下的所有Go文件进行静态代码检查
$ golangci-lint run dir1 dir2/... dir3/file1.go # 运行指定的 errcheck linter# 对指定的Go文件或者指定目录下的Go文件进行静态代码检查
$ golangci-lint run -c .golangci.yaml ./... # 根据指定配置文件，进行静态代码检查
$ golangci-lint run --no-config --disable-all -E errcheck ./... # 运行指定的 errcheck linter
$ golangci-lint run --no-config -D godot,errcheck # 禁止运行指定的godot,errcheck liner
```

此外，golangci-lint 还支持 //nolint 、//nolint:golint,unused 等方式来减少误报。 最后，分享了一些自己使用 golangci-lint 时总结的经验。例如：

- 第一次修改，可以按目录修改；
- 按文件修改，减少文件切换次数，提高修改效率；
- 尽可能多地使用 golangci-lint 提供的 linter。

希望这些建议对使用 golangci-lint 有一定帮助。 

### 课后练习

- 执行golangci-lint linters命令，查看 golangci-lint 支持哪些 linter，以及这些 linter 的作用。
- 思考下，如何在 golangci-lint 中集成自定义的 linter。



## Go 项目之 API 文档

作为一名开发者，通常讨厌编写文档，因为这是一件重复和缺乏乐趣的事情。但是在 开发过程中，又有一些文档是必须要编写的，比如 API 文档。 

一个企业级的 Go 后端项目，通常也会有个配套的前端。为了加快研发进度，通常是后端和前端并行开发，这就需要后端开发者在开发后端代码之前，先设计好 API 接口，提供给前端。所以在设计阶段，就需要生成 API 接口文档。 

一个好的 API 文档，可以减少用户上手的复杂度，也意味着更容易留住用户。好的 API 文档也可以减少沟通成本，帮助开发者更好地理解 API 的调用方式，从而节省时间，提高开发效率。

这时候，一定希望有一个工具能够自动生成 API 文档，解放双手。Swagger 就是这么一个工具，可以生成易于共享且具有足够描述性的 API 文档。 

接下来，就来看下，如何使用 Swagger 生成 API 文档。 

### Swagger 介绍 

Swagger 是一套围绕 OpenAPI 规范构建的开源工具，可以设计、构建、编写和使用 REST API。

Swagger 包含很多工具，其中主要的 Swagger 工具包括：

- Swagger 编辑器：基于浏览器的编辑器，可以在其中编写 OpenAPI 规范，并实时预览 API 文档。https://editor.swagger.io 就是一个 Swagger 编辑器，可以尝试在其中编辑和预览 API 文档。 
- Swagger UI：将 OpenAPI 规范呈现为交互式 API 文档，并可以在浏览器中尝试 API 调用。 
- Swagger Codegen：根据 OpenAPI 规范，生成服务器存根和客户端代码库，目前已涵盖了 40 多种语言。

### Swagger 和 OpenAPI 的区别 

在谈到 Swagger 时，也经常会谈到 OpenAPI。那么二者有什么区别呢？ 

OpenAPI 是一个 API 规范，它的前身叫 Swagger 规范，通过定义一种用来描述 API 格式 或 API 定义的语言，来规范 RESTful 服务开发过程，目前最新的 OpenAPI 规范是 OpenAPI 3.0（也就是 Swagger 2.0 规范）。 

OpenAPI 规范规定了一个 API 必须包含的基本信息，这些信息包括：

- 对 API 的描述，介绍 API 可以实现的功能。 
- 每个 API 上可用的路径（/users）和操作（GET /users，POST /users）。 
- 每个 API 的输入 / 返回的参数。 
- 验证方法。 
- 联系信息、许可证、使用条款和其他信息。

所以，可以简单地这么理解：OpenAPI 是一个 API 规范，Swagger 则是实现规范的工具。 

另外，要编写 Swagger 文档，首先要会使用 Swagger 文档编写语法，因为语法比较多， 这里就不多介绍了，可以参考 Swagger 官方提供的 OpenAPI Specification 来学习。 

### 用 go-swagger 来生成 Swagger API 文档 

在 Go 项目开发中，可以通过下面两种方法来生成 Swagger API 文档： 

- 第一，如果熟悉 Swagger 语法的话，可以直接编写 JSON/YAML 格式的 Swagger 文档。建议选择 YAML 格式，因为它比 JSON 格式更简洁直观。 
- 第二，通过工具生成 Swagger 文档，目前可以通过 swag和 go-swagger两个工具来生成。

对比这两种方法，直接编写 Swagger 文档，不比编写 Markdown 格式的 API 文档工作量小，不符合程序员“偷懒”的习惯。所以，本专栏就使用 go-swagger 工具， 基于代码注释来自动生成 Swagger 文档。

为什么选 go-swagger 呢？有这么几个原因：

- go-swagger 比 swag 功能更强大：go-swagger 提供了更灵活、更多的功能来描述 API。 
- 使代码更易读：如果使用 swag，每一个 API 都需要有一个冗长的注释，有时候代码注释比代码还要长，但是通过 go-swagger 可以将代码和注释分开编写， 一方面可以使代码保持简洁，清晰易读，另一方面可以在另外一个包中，统一管理这些 Swagger API 文档定义。 
- 更好的社区支持：go-swagger 目前有非常多的 Github star 数，出现 Bug 的概率很小，并且处在一个频繁更新的活跃状态。

已经知道了，go-swagger 是一个功能强大的、高性能的、可以根据代码注释生成 Swagger API 文档的工具。除此之外，go-swagger 还有很多其他特性：

- 根据 Swagger 定义文件生成服务端代码。 
- 根据 Swagger 定义文件生成客户端代码。
- 校验 Swagger 定义文件是否正确。 
- 启动一个 HTTP 服务器，可以通过浏览器访问 API 文档。 
- 根据 Swagger 文档定义的参数生成 Go model 结构体定义。

可以看到，使用 go-swagger 生成 Swagger 文档，可以减少编写文档的时间， 提高开发效率，并能保证文档的及时性和准确性。 

这里需要注意，如果要对外提供 API 的 Go SDK，可以考虑使用 go-swagger 来生成 客户端代码。但是 go-swagger 生成的服务端代码不够优雅，所以建议自行编写 服务端代码。 

目前，有很多知名公司和组织的项目都使用了 go-swagger，例如 Moby、CoreOS、 Kubernetes、Cilium 等。 

### 安装 Swagger 工具 

go-swagger 通过 swagger 命令行工具来完成其功能，swagger 安装方法如下：

```sh
$ go get -u github.com/go-swagger/go-swagger/cmd/swagger
$ swagger version
version: v0.28.0
commit: (unknown, mod sum: "h1:cFzm/DrsqKiDeBpzRDu5N3vjraU3O9IfpFfz+TscKWY=")
```

### swagger 命令行工具介绍 

swagger 命令格式为 swagger [OPTIONS] 。可以通过 `swagger -h` 查看 swagger 使用帮助。swagger 提供的子命令及功能见下表：

![image-20211118202908051](IAM-document.assets/image-20211118202908051.png)

### 如何使用 swagger 命令生成 Swagger 文档？ 

go-swagger 通过解析源码中的注释来生成 Swagger 文档，go-swagger 的详细注释语法可参考官方文档。常用的有如下几类注释语法：

![image-20211118203021802](IAM-document.assets/image-20211118203021802.png)

### 解析注释生成 Swagger 文档 

swagger generate 命令会找到 main 函数，然后遍历所有源码文件，解析源码中与 Swagger 相关的注释，然后自动生成 swagger.json/swagger.yaml 文件。

#### main.go 

这一过程的示例代码为 gopractise-demo/swagger。目录下有一个 main.go 文件，定义了如下 API 接口：

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/marmotedu/gopractise-demo/swagger/api"
	// This line is necessary for go-swagger to find your docs!
	_ "github.com/marmotedu/gopractise-demo/swagger/docs"
)

var users []*api.User

func main() {
	r := gin.Default()
	r.POST("/users", Create)
	r.GET("/users/:name", Get)

	log.Fatal(r.Run(":5555"))
}

// Create create a user in memory.
func Create(c *gin.Context) {
	var user api.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error(), "code": 10001})
		return
	}

	for _, u := range users {
		if u.Name == user.Name {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("user %s already exist", user.Name), "code": 10001})
			return
		}
	}

	users = append(users, &user)
	c.JSON(http.StatusOK, user)
}

// Get return the detail information for a user.
func Get(c *gin.Context) {
	username := c.Param("name")
	for _, u := range users {
		if u.Name == username {
			c.JSON(http.StatusOK, u)
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("user %s not exist", username), "code": 10002})
}
```

#### user.go

main 包中引入的 User struct 位于 gopractise-demo/swagger/api 目录下的user.go 文件：

```go
// Package api defines the user model.
package api

// User represents body of User request and response.
type User struct {
	// User's name.
	// Required: true
	Name string `json:"name"`

	// User's nickname.
	// Required: true
	Nickname string `json:"nickname"`

	// User's address.
	Address string `json:"address"`

	// User's email.
	Email string `json:"email"`
}
```

// Required: true 说明字段是必须的，生成 Swagger 文档时，也会在文档中声明该 字段是必须字段。 

为了使代码保持简洁，在另外一个 Go 包中编写带 go-swagger 注释的 API 文档。假 设该 Go 包名字为 docs，在开始编写 Go API 注释之前，需要在 main.go 文件中导入 docs 包：

```go
_ "github.com/marmotedu/gopractise-demo/swagger/docs"
```

通过导入 docs 包，可以使 go-swagger 在递归解析 main 包的依赖包时，找到 docs 包，并解析包中的注释。

在 gopractise-demo/swagger 目录下，创建 docs 文件夹：

```sh
$ mkdir docs
$ cd docs
```

#### doc.go

在 docs 目录下，创建一个 doc.go 文件，在该文件中提供 API 接口的基本信息：

```go
// Package docs awesome.
//
// Documentation of our awesome API.
//
//     Schemes: http, https
//     BasePath: /
//     Version: 0.1.0
//     Host: some-url.com
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - basic
//
//    SecurityDefinitions:
//    basic:
//      type: basic
//
// swagger:meta
package docs
```

Package docs 后面的字符串 awesome 代表我们的 HTTP 服务名。Documentation of our awesome API 是我们 API 的描述。其他都是 go-swagger 可识别的注释，代表一定的意义。最后以 swagger:meta 注释结束。 

#### 生成 YAML 格式 Swagger API 文档

编写完 doc.go 文件后，进入 gopractise-demo/swagger 目录，执行如下命令，生成 Swagger API 文档，并启动 HTTP 服务，在浏览器查看 Swagger：

```sh
$ swagger generate spec -o swagger.yaml
$ swagger serve --no-open -F=swagger --port 36666 swagger.yaml
2021/11/18 20:48:39 serving docs at http://localhost:36666/docs
```

- -o：指定要输出的文件名。swagger 会根据文件名后缀.yaml 或者.json，决定生成的文件格式为 YAML 或 JSON。 
- --no-open：因为是在 Linux 服务器下执行命令，没有安装浏览器，所以使 --no-open 禁止调用浏览器打开 URL。
-  -F：指定文档的风格，可选 swagger 和 redoc。选用了 redoc，因为觉得 redoc 格式更加易读和清晰。
- --port：指定启动的 HTTP 服务监听端口。

打开浏览器，访问 http://localhost:36666/docs ，如下图所示：

![image-20211118205140133](IAM-document.assets/image-20211118205140133.png)

#### 生成 JSON 格式 Swagger API 文档

如果想要 JSON 格式的 Swagger 文档，可执行如下命令，将生成的 swagger.yaml 转换为 swagger.json：

```sh
$ swagger generate spec -i ./swagger.yaml -o ./swagger.json
```

#### API 接口的定义文件

接下来，就可以编写 API 接口的定义文件（位于gopractise-demo/swagger/docs/user.go文件中）：

```go
package docs

import (
	"github.com/marmotedu/gopractise-demo/swagger/api"
)

// swagger:route POST /users user createUserRequest
// Create a user in memory.
// responses:
//   200: createUserResponse
//   default: errResponse

// swagger:route GET /users/{name} user getUserRequest
// Get a user from memory.
// responses:
//   200: getUserResponse
//   default: errResponse

// swagger:parameters createUserRequest
type userParamsWrapper struct {
	// This text will appear as description of your request body.
	// in:body
	Body api.User
}

// This text will appear as description of your request url path.
// swagger:parameters getUserRequest
type getUserParamsWrapper struct {
	// in:path
	Name string `json:"name"`
}

// This text will appear as description of your response body.
// swagger:response createUserResponse
type createUserResponseWrapper struct {
	// in:body
	Body api.User
}

// This text will appear as description of your response body.
// swagger:response getUserResponse
type getUserResponseWrapper struct {
	// in:body
	Body api.User
}

// This text will appear as description of your error response body.
// swagger:response errResponse
type errResponseWrapper struct {
	// Error code.
	Code int `json:"code"`

	// Error message.
	Message string `json:"message"`
}
```

user.go 文件说明：

- swagger:route：swagger:route代表 API 接口描述的开始，后面的字符串格式为 HTTP方法 URL Tag ID。
  - 可以填写多个 tag，相同 tag 的 API 接口在 Swagger 文档中会被分为一组。
  - ID 是一个标识符，swagger:parameters 是具有相同 ID 的 swagger:route的请求参数。
- swagger:route下面的一行是该 API 接口的描述，需要以英文点号为结尾。
- responses:定义了 API 接口的返回参数，例如当 HTTP 状态码 是 200 时，返回 createUserResponse，createUserResponse 会跟 swagger:response进行匹配，匹配成功的 swagger:response 就是该 API 接口返回 200 状态码时的返回。 
- swagger:response：swagger:response定义了 API 接口的返回，例如 getUserResponseWrapper，关于名字，可以根据需要自由命名，并不会带来任何不同。
- getUserResponseWrapper 中有一个 Body 字段，其注释为// in:body，说明该参数是在 HTTP Body 中返回。
- swagger:response之上的注释会被解析为返回参 数的描述。
- api.User 自动被 go-swagger 解析为 Example Value和Model。不用再去编写重复的返回字段，只需要引用已有的 Go 结构体即可，这也是通过工具生成 Swagger 文档的魅力所在。 
- swagger:parameters：swagger:parameters定义了 API 接口的请求参数，例如 userParamsWrapper。
- userParamsWrapper 之上的注释会被解析为请求参数的描述，// in:body代表该参数是位于 HTTP Body 中。
- 同样，userParamsWrapper 结构体名也可以随意命名，不会带来任何不同。
- swagger:parameters之后的 createUserRequest 会跟swagger:route的 ID 进行匹配，匹配成功则说明是该 ID 所在 API 接口的请求参数。

#### 浏览器查看 Swagger

进入 gopractise-demo/swagger 目录，执行如下命令，生成 Swagger API 文档，并启 动 HTTP 服务，在浏览器查看 Swagger：

```sh
$ swagger generate spec -o swagger.yaml
$ swagger serve --no-open -F=swagger --port 36666 swagger.yaml
# or
$ swagger serve --no-open -F=redoc --port 36666 swagger.yaml
2021/11/18 20:48:39 serving docs at http://localhost:36666/docs
```

打开浏览器，访问  http://localhost:36666/docs ，如下图所示：

![image-20211118210701107](IAM-document.assets/image-20211118210701107.png)

上面生成了 swagger 风格的 UI 界面，也可以使用 redoc 风格的 UI 界面，如下图所示：

![image-20211118210853514](IAM-document.assets/image-20211118210853514.png)

### go-swagger 其他功能介绍 

上面，介绍了 swagger 最常用的 generate、serve 命令，关于 swagger 其他有用的命令，这里也简单介绍一下。

#### 对比 Swagger 文档

1. 对比 Swagger 文档

```sh
$ swagger diff -d change.log swagger.new.yaml swagger.old.yaml
$ cat change.log
BREAKING CHANGES:
=================
/users:post Request - Body.Body.nickname.address.email.name.Body : User - Dele
compatibility test FAILED: 1 breaking changes detected
```

#### 生成服务端代码

2. 生成服务端代码

也可以先定义 Swagger 接口文档，再用 swagger 命令，基于 Swagger 接口文档生 成服务端代码。

假设应用名为 go-user，进入 gopractise-demo/swagger 目录， 创建 go-user 目录，并生成服务端代码：

```sh
$ mkdir go-user
$ cd go-user
$ swagger generate server -f ../swagger.yaml -A go-user
```

上述命令会在当前目录生成 cmd、restapi、models 文件夹，可执行如下命令查看 server 组件启动方式：

```sh
$ go run cmd/go-user-server/main.go -h
```

#### 生成客户端代码

3. 生成客户端代码

在 go-user 目录下执行如下命令：

```sh
$ swagger generate client -f ../swagger.yaml -A go-user
```

上述命令会在当前目录生成 client，包含了 API 接口的调用函数，也就是 API 接口的 Go SDK。

#### 验证 Swagger 文档是否合法

4. 验证 Swagger 文档是否合法

```sh
$ swagger validate swagger.yaml
2021/11/18 21:44:34 
The swagger spec at "swagger.yaml" is valid against swagger specification 2.0
```

#### 合并 Swagger 文档

5. 合并 Swagger 文档

```sh
$ swagger mixin swagger_part1.yaml swagger_part2.yaml
```

### IAM Swagger 文档 

IAM 的 Swagger 文档定义在 iam/api/swagger/docs 目录下，遵循 go-swagger 规范进行定义。 

iam/api/swagger/docs/doc.go文件定义了更多 Swagger 文档的基本信息，比如开源协议、联系方式、安全认证等。 

更详细的定义，可以直接查看 iam/api/swagger/docs 目录下的 Go 源码文件。 

为了便于生成文档和启动 HTTP 服务查看 Swagger 文档，该操作被放在 Makefile 中执行 （位于 iam/scripts/make-rules/swagger.mk文件中）：

```makefile
.PHONY: swagger.run
swagger.run: tools.verify.swagger
	@echo "===========> Generating swagger API docs"
	@swagger generate spec --scan-models -w $(ROOT_DIR)/cmd/genswaggertypedocs -o $(ROOT_DIR)/api/swagger/swagger.yaml

.PHONY: swagger.serve
swagger.serve: tools.verify.swagger
	@swagger serve -F=redoc --no-open --port 36666 $(ROOT_DIR)/api/swagger/swagger.yaml
```

Makefile 文件说明：

- tools.verify.swagger：检查 Linux 系统是否安装了 go-swagger 的命令行工具 swagger，如果没有安装则运行 go get 安装。 
- swagger.run：运行 swagger generate spec 命令生成 Swagger 文档 swagger.yaml，运行前会检查 swagger 是否安装。 
  - --scan-models 指定生成的文档中包含带有 swagger:model 注释的 Go Models。
  - -w 指定 swagger 命令运行的目录。 
- swagger.serve：运行 swagger serve 命令生成 Swagger 文档 swagger.yaml，运行前会检查 swagger 是否安装。

在 iam 源码根目录下执行如下命令，即可生成并启动 HTTP 服务查看 Swagger 文档：

```sh
$ make swagger
$ make serve-swagger
2021/11/18 21:58:15 serving docs at http://localhost:36666/docs
```

打开浏览器，打开 http://x.x.x.x:36666/docs 查看 Swagger 文档，x.x.x.x 是服务器 的 IP 地址，如下图所示：

![image-20211118220146053](IAM-document.assets/image-20211118220146053.png)

IAM 的 Swagger 文档，还可以通过在 iam 源码根目录下执行 go generate ./...命令 生成，为此，需要在 iam/cmd/genswaggertypedocs/swagger_type_docs.go 文件 中，添加 //go:generate 注释。如下图所示：

![image-20211118220328587](IAM-document.assets/image-20211118220328587.png)

### 总结 

在做 Go 服务开发时，要向前端或用户提供 API 文档，手动编写 API 文档工作量大， 也难以维护。所以，现在很多项目都是自动生成 Swagger 格式的 API 文档。

提到 Swagger，很多开发者不清楚其和 OpenAPI 的区别，所以总结了：OpenAPI 是一个 API 规范，Swagger 则是实现规范的工具。 

在 Go 中，用得最多的是利用 go-swagger 来生成 Swagger 格式的 API 文档。go-swagger 包含了很多语法，可以访问 Swagger 2.0进行学习。学习完 Swagger 2.0 的语法之后，就可以编写 swagger 注释了，之后可以通过

```sh
$ swagger generate spec -o swagger.yaml
```

来生成 swagger 文档 swagger.yaml。通过

```sh
$ swagger serve --no-open -F=swagger --port 36666 swagger.yaml
```

来提供一个前端界面，供访问 swagger 文档。 

为了方便管理，可以将 swagger generate spec 和 swagger serve 命令加入到 Makefile 文件中，通过 Makefile 来生成 Swagger 文档，并提供给前端界面。 

### 课后练习

- 尝试将当前项目的一个 API 接口，用 go-swagger 生成 swagger 格式的 API 文档。 
- 思考下，为什么 IAM 项目的 swagger 定义文档会放在 iam/api/swagger/docs 目录下，这样做有什么好处？
  - 放在api目录下，说明这个是api的定义文件。API文档聚合在一 个目录下，后期维护，查看都很方便。



## Go 项目之设计业务的错误码

如何设计业务的错误码。 

现代的软件架构，很多都是对外暴露 RESTful API 接口，内部系统通信采用 RPC 协议。因为 RESTful API 接口有一些天生的优势，比如规范、调试友好、易懂，所以通常作为直接 面向用户的通信规范。 

既然是直接面向用户，那么首先就要求消息返回格式是规范的；其次，如果接口报错，还要能给用户提供一些有用的报错信息，通常需要包含 Code 码（用来唯一定位一次错误） 和 Message（用来展示出错的信息）。这就需要设计一套规范的、科学的错误码。 

这一讲，就来详细介绍下，如何设计一套规范的、科学的错误码。下一讲，我还会介绍如何提供一个 errors 包来支持设计的错误码。

### 期望错误码实现的功能 

要想设计一套错误码，首先就得弄清需求。 

RESTful API 是基于 HTTP 协议的一系列 API 开发规范，HTTP 请求结束后，无论 API 请求成功或失败，都需要让客户端感知到，以便客户端决定下一步该如何处理。

为了让用户拥有最好的体验，需要有一个比较好的错误码实现方式。这里介绍下在设计错误码时，期望能够实现的功能。 

#### 有业务 Code 码标识

第一个功能是有业务 Code 码标识。 

因为 HTTP Code 码有限，并且都是跟 HTTP Transport 层相关的 Code 码，所以希望能有自己的错误 Code 码。一方面，可以根据需要自行扩展，另一方面也能够精准地定位到具体是哪个错误。

同时，因为 Code 码通常是对计算机友好的 10 进制整数，基于 Code 码，计算机也可以很方便地进行一些分支处理。当然了，业务码也要有一定规则，可以通过业务码迅速定位出是哪类错误。 

#### 对外对内分别展示不同的错误信息

第二个功能，考虑到安全，希望能够对外对内分别展示不同的错误信息。 

当开发一个对外的系统，业务出错时，需要一些机制告诉用户出了什么错误，如果能够提供一些帮助文档会更好。但是，不可能把所有的错误都暴露给外部用户，这不仅没必要，也不安全。

所以也需要能让获取到更详细的内部错误信息的机制，这些内部错误 信息可能包含一些敏感的数据，不宜对外展示，但可以协助进行问题定位。

 所以，需要设计的错误码应该是规范的，能方便客户端感知到 HTTP 是否请求成功， 并带有业务码和出错信息。

### 常见的错误码设计方式 

在业务中，大致有三种错误码实现方式。

用一次因为用户账号没有找到而请求失败的例子，分别解释一下： 

#### 返回 200 http status code

第一种方式，不论请求成功或失败，始终返回 200 http status code，在 HTTP Body 中包含用户账号没有找到的错误信息。

例如 Facebook API 的错误 Code 设计，始终返回 200 http status code：

```json
{
  "error": {
    "message": "Syntax error \"Field picture specified more than once. This is only possible before version 2.1\" at character 23: id,name,picture,picture"
    "type": "OAuthException",
    "code": 2500,
    "fbtrace_id": "xxxxxxxxxxx"
  }
}
```

采用固定返回 200 http status code的方式，有其合理性。比如，HTTP Code 通常代表 HTTP Transport 层的状态信息。当收到 HTTP 请求，并返回时，HTTP Transport 层是成功的，所以从这个层面上来看，HTTP Status 固定为 200 也是合理的。 

但是这个方式的缺点也很明显：对于每一次请求，都要去解析 HTTP Body，从中解析出错误码和错误信息。实际上，大部分情况下，对于成功的请求，要么直接转发，要么直接解析到某个结构体中；对于失败的请求，也希望能够更直接地感知到请求失败。

这种方式对性能会有一定的影响，对客户端不友好。所以不建议使用这种方式。 

#### 返回http 404 Not Found错误码,返回简单的错误信息

第二种方式，返回 http 404 Not Found错误码，并在 Body 中返回简单的错误信息。 

例如：Twitter API 的错误设计，会根据错误类型，返回合适的 HTTP Code，并在 Body 中返回错误信息和自定义业务 Code。

```http
HTTP/1.1 400 Bad Request
x-connection-hash: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
set-cookie: guest_id=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Date: Thu, 01 Jun 2017 03:04:23 GMT
Content-Length: 62
x-response-time: 5
strict-transport-security: max-age=631138519
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Server: tsa_b

{"errors":[{"code":215,"message":"Bad Authentication data."}]}
```

这种方式比第一种要好一些，通过http status code可以使客户端非常直接地感知到请求失败，并且提供给客户端一些错误信息供参考。但是仅仅靠这些信息，还不能准确地定位和解决问题。 

#### 返回http 404 Not Found错误码,返回详细的错误信息

第三种方式，返回http 404 Not Found错误码，并在 Body 中返回详细的错误信息。 

例如：微软 Bing API 的错误设计，会根据错误类型，返回合适的 HTTP Code，并在 Body 中返回详尽的错误信息。

```http
HTTP/1.1 400
Date: Thu, 01 Jun 2017 03:40:55 GMT
Content-Length: 276
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Server: Microsoft-IIS/10.0
X-Content-Type-Options: nosniff

{"SearchResponse":{"Version":"2.2","Query":{"SearchTerms":"api error codes"},"
```

这是比较推荐的一种方式，既能通过http status code使客户端方便地知道请求出 错，又可以使用户根据返回的信息知道哪里出错，以及如何解决问题。

同时，返回了机器友好的业务 Code 码，可以在有需要时让程序进一步判断处理。

### 错误码设计建议 

#### 优秀的错误码设计思路

综合刚才讲到的，可以总结出一套优秀的错误码设计思路：

- 有区别于 http status code 的业务码，业务码需要有一定规则，可以通过业务码判断出是哪类错误。 
- 请求出错时，可以通过http status code直接感知到请求出错。 
- 需要在请求出错时，返回详细的信息，通常包括 3 类信息：业务 Code 码、错误信息和参考文档（可选）。 
- 返回的错误信息，需要是可以直接展示给用户的安全信息，也就是说不能包含敏感信 息；同时也要有内部更详细的错误信息，方便 debug。
- 返回的数据格式应该是固定的、规范的。 
- 错误信息要保持简洁，并且提供有用的信息。

这里其实还有两个功能点需要实现：业务 Code 码设计，以及请求出错时，如何设置 http status code。 

接下来，详细介绍下如何实现这两个功能点。 

#### 业务 Code 码设计 

要解决业务 Code 码如何设计这个问题，先来看下为什么要引入业务 Code 码。 

在实际开发中，引入业务 Code 码有下面几个好处：

- 可以非常方便地定位问题和定位代码行（看到错误码知道什么意思、grep 错误码可以定位到错误码所在行、某个错误类型的唯一标识）。 
- 错误码包含一定的信息，通过错误码可以判断出错误级别、错误模块和具体错误信息。 
- Go 中的 HTTP 服务器开发都是引用 net/http 包，该包中只有 60 个错误码，基本都是跟 HTTP 请求相关的错误码，在一个大型系统中，这些错误码完全不够用，而且这些错误码跟业务没有任何关联，满足不了业务的需求。引入业务的 Code 码，则可以解决这些问题。 
- 业务开发过程中，可能需要判断错误是哪种类型，以便做相应的逻辑处理，通过定制的错误可以很容易做到这点，例如：

```go
if err == code.ErrBind {
	...
}
```

这里要注意，业务 Code 码可以是一个整数，也可以是一个整型字符串，还可以是一个字符型字符串，它是错误的唯一标识。 

通过研究腾讯云、阿里云、新浪的开放 API，发现新浪的 API Code 码设计更合理些。 所以，参考新浪的 Code 码设计，总结出推荐的 Code 码设计规范：纯数字表示，不同部位代表不同的服务，不同的模块。 

错误代码说明：100101

- 10: 服务。 
- 01: 某个服务下的某个模块。 
- 01: 模块下的错误码序号，每个模块可以注册 100 个错误

通过100101可以知道这个错误是服务 A，数据库模块下的记录没有找到错误。 

可能会问：按这种设计，每个模块下最多能注册 100 个错误，是不是有点少？其实,如果每个模块的错误码超过 100 个，要么说明这个模块太大了，建议拆分；要么说 明错误码设计得不合理，共享性差，需要重新设计。 

#### 如何设置 HTTP Status Code 

Go net/http 包提供了 60 个错误码，大致分为如下 5 类：

- 1XX - （指示信息）表示请求已接收，继续处理。 
- 2XX - （请求成功）表示成功处理了请求的状态代码。 
- 3XX - （请求被重定向）表示要完成请求，需要进一步操作。通常，这些状态代码用来重定向。 
- 4XX - （请求错误）这些状态代码表示请求可能出错，妨碍了服务器的处理，通常是客户端出错，需要客户端做进一步的处理。 
- 5XX - （服务器错误）这些状态代码表示服务器在尝试处理请求时发生内部错误。这些错误可能是服务器本身的错误，而不是客户端的问题。

可以看到 HTTP Code 有很多种，如果每个 Code 都做错误映射，会面临很多问题。比如，研发同学不太好判断错误属于哪种http status code，到最后很可能会导致错误或者http status code不匹配，变成一种形式。

而且，客户端也难以应对这么多的 HTTP 错误码。 

所以，这里建议http status code不要太多，基本上只需要这 3 个 HTTP Code:

- 200 - 表示请求成功执行。 
- 400 - 表示客户端出问题。 
- 500 - 表示服务端出问题。

如果觉得这 3 个错误码不够用，最多可以加如下 3 个错误码：

- 401 - 表示认证失败。
- 403 - 表示授权失败。 
- 404 - 表示资源找不到，这里的资源可以是 URL 或者 RESTful 资源。

将错误码控制在适当的数目内，客户端比较容易处理和判断，开发也比较容易进行错误码映射。 

### IAM 项目错误码设计规范 

接下来，来看下 IAM 项目的错误码是如何设计的。 

#### Code 设计规范 

先来看下 IAM 项目业务的 Code 码设计规范，具体实现可参考 internal/pkg/code 目 录。IAM 项目的错误码设计规范符合上面介绍的错误码设计思路和规范，具体规范见下。 

Code 代码从 100101 开始，1000 以下为 github.com/marmotedu/errors 保留 code。 

错误代码说明：100101

![image-20211119002433376](IAM-document.assets/image-20211119002433376.png)

#### 服务和模块说明

通用：说明所有服务都适用的错误，提高复用性，避免重复造轮子。

![image-20211119003018544](IAM-document.assets/image-20211119003018544.png)

#### 错误信息规范说明

- 对外暴露的错误，统一大写开头，结尾不要加.。 
- 对外暴露的错误要简洁，并能准确说明问题。 
- 对外暴露的错误说明，应该是 该怎么做 而不是哪里错了。

这里需要注意，错误信息是直接暴露给用户的，不能包含敏感信息。 

#### IAM API 接口返回值说明 

如果返回结果中存在 code 字段，则表示调用 API 接口失败。例如：

```json
{
  "code": 100101,
  "message": "Database error",
  "reference": "https://github.com/marmotedu/iam/tree/master/docs/guide/zh-CN/faq/iam-apiserver"
}
```

上述返回中 code 表示错误码，message 表示该错误的具体信息。每个错误同时也对应一个 HTTP 状态码。

比如上述错误码对应了 HTTP 状态码 500(Internal Server Error)。另外，在出错时，也返回了reference字段，该字段包含了可以解决这个错误的文档链接地址。 

关于 IAM 系统支持的错误码，列了一个表格，可以看看：

![image-20211119003624752](IAM-document.assets/image-20211119003624752.png)

![image-20211119003726126](IAM-document.assets/image-20211119003726126.png)

### 总结 

对外暴露的 API 接口需要有一套规范的、科学的错误码。目前业界的错误码大概有 3 种设计方式，用一次因为用户账号没有找到而请求失败的例子，做了解释：

- 不论请求成功失败，始终返回200 http status code，在 HTTP Body 中包含用户 账号没有找到的错误信息。
- 返回http 404 Not Found错误码，并在 Body 中返回简单的错误信息。 
- 返回http 404 Not Found错误码，并在 Body 中返回详细的错误信息。

这一讲，参考这 3 个错误码设计，给出了错误码设计建议：

- 错误码包含 HTTP Code 和业务 Code，并且业务 Code 会映射为一个 HTTP Code。
- 错误码也会对外暴露两种错误信息，一种是直接暴露给用户的，不包含敏感信息的信息；另一种是供内部开发查看，定位问题的错误信息。
- 该错误码还支持返回参考文档，用于在出错时展示给用户，供 用户查看解决问题。 

重点关注 Code 码设计规范：纯数字表示，不同部位代表不同的服务，不 =同的模块。 

比如错误代码100101，其中 10 代表服务；中间的 01 代表某个服务下的某个模块；最后 的 01 代表模块下的错误码序号，每个模块可以注册 100 个错误。 

### 课后练习

- 既然错误码是符合规范的，请思考下：有没有一种 Low Code 的方式，根据错误码规范，自动生成错误码文档呢？ 
- 思考下还遇到过哪些更科学的错误码设计。



## Go 项目之设计业务的错误包

在 Go 项目开发中，错误是必须要处理的一个事项。除了上一讲学习过的错误码，处理错误也离不开错误包。 

业界有很多优秀的、开源的错误包可供选择，例如 Go 标准库自带的errors包、 github.com/pkg/errors包。但是这些包目前还不支持业务错误码，很难满足生产级应用的需求。

所以，在实际开发中，有必要开发出适合自己错误码设计的错误包。当然，也没必要自己从 0 开发，可以基于一些优秀的包来进行二次封装。 

这一讲里，就来一起看看，如何设计一个错误包来适配上一讲设计的错误码，以及一个错误码的具体实现。

### 错误包需要具有哪些功能？ 

要想设计一个优秀的错误包，首先得知道一个优秀的错误包需要具备哪些功能。至少需要有下面这六个功能： 

#### 支持错误堆栈

首先，应该能支持错误堆栈。来看下面一段代码，假设保存在 bad.go文件中：

```go
package main

import (
	"fmt"
	"log"
)

func main() {
	if err := funcA(); err != nil {
		log.Fatalf("call func got failed: %v", err)
		return
	}
	log.Println("call func success")
}
func funcA() error {
	if err := funcB(); err != nil {
		return err
	}
	return fmt.Errorf("func called error")
}
func funcB() error {
	return fmt.Errorf("func called error")
}
```

执行上面的代码：

```sh
$ go run bad.go
2021/11/19 21:27:16 call func got failed: func called error
exit status 1
```

这时想定位问题，但不知道具体是哪行代码报的错误，只能靠猜，还不一定能猜到。 

为了解决这个问题，可以加一些 Debug 信息，来协助定位问题。这样做在测试环境是没问题的，但是在线上环境，一方面修改、发布都比较麻烦，另一方面问题可能比较难重现。

这时候会想，要是能打印错误的堆栈就好了。例如：

```sh
2021/07/02 14:17:03 call func got failed: func called error
main.funcB
/home/colin/workspace/golang/src/github.com/marmotedu/gopractise-demo/errors
main.funcA
/home/colin/workspace/golang/src/github.com/marmotedu/gopractise-demo/errors
main.main
/home/colin/workspace/golang/src/github.com/marmotedu/gopractise-demo/errors
runtime.main
/home/colin/go/go1.16.2/src/runtime/proc.go:225
runtime.goexit
/home/colin/go/go1.16.2/src/runtime/asm_amd64.s:1371
exit status 1
```

通过上面的错误输出，可以很容易地知道是哪行代码报的错，从而极大提高问题定位的效率，降低定位的难度。所以，一个优秀的 errors 包，首先需要支持错误堆栈。 

#### 支持不同的打印格式

其次，能够支持不同的打印格式。例如%+v、%v、%s等格式，可以根据需要打印不同丰富度的错误信息。 

#### 支持 Wrap/Unwrap 功能

再次，能支持 Wrap/Unwrap 功能，也就是在已有的错误上，追加一些新的信息。

例如 errors.Wrap(err, "open file failed") 。Wrap 通常用在调用函数中，调用函数可以基于被调函数报错时的错误 Wrap 一些自己的信息，丰富报错信息，方便后期的错误定位，例如：

```go
func funcA() error {
	if err := funcB(); err != nil {
		return errors.Wrap(err, "call funcB failed")
	}
	return errors.New("func called error")
}
func funcB() error {
	return errors.New("func called error")
}
```

这里要注意，不同的错误类型，Wrap 函数的逻辑也可以不同。另外，在调用 Wrap 时， 也会生成一个错误堆栈节点。既然能够嵌套 error，那有时候还可能需要获取被嵌套的 error，这时就需要错误包提供Unwrap函数。 

#### 有Is方法

还有，错误包应该有Is方法。在实际开发中，经常需要判断某个 error 是否是指定的 error。

在 Go 1.13 之前，也就是没有 wrapping error 的时候，要判断 error 是不是同一个，可以使用如下方法：

```go
if err == os.ErrNotExist {
	// normal code
}
```

但是现在，因为有了 wrapping error，这样判断就会有问题。因为根本不知道返回的 err 是不是一个嵌套的 error，嵌套了几层。这种情况下，错误包就需要提供Is函数：

```go
func Is(err, target error) bool
```

当 err 和 target 是同一个，或者 err 是一个 wrapping error 的时候，如果 target 也包含在这个嵌套 error 链中，返回 true，否则返回 fasle。 

#### 支持 As 函数

另外，错误包应该支持 As 函数。 

在 Go 1.13 之前，没有 wrapping error 的时候，要把 error 转为另外一个 error，一般都是使用 type assertion 或者 type switch，也就是类型断言。例如：

```go
if perr, ok := err.(*os.PathError); ok {
    fmt.Println(perr.Path)
}
```

但是现在，返回的 err 可能是嵌套的 error，甚至好几层嵌套，这种方式就不能用了。所以，可以通过实现 As 函数来完成这种功能。

现在把上面的例子，用 As 函数实现一下：

```go
var perr *os.PathError
if errors.As(err, &perr) {
  fmt.Println(perr.Path)
}
```

这样就可以完全实现类型断言的功能，而且还更强大，因为它可以处理 wrapping error。 

#### 支持两种错误创建方式

最后，能够支持两种错误创建方式：非格式化创建和格式化创建。例如：

```go
errors.New("file not found")
errors.Errorf("file %s not found", "iam-apiserver")
```

上面，介绍了一个优秀的错误包应该具备的功能。一个好消息是，Github 上有不少实现了这些功能的错误包，其中github.com/pkg/errors 包最受欢迎。

所以，基于 github.com/pkg/errors 包进行了二次封装，用来支持上一讲所介绍的错误码。 

### 错误包实现 

明确优秀的错误包应该具备的功能后，来看下错误包的实现。实现的源码存放在 github.com/marmotedu/errors。

#### withCode 结构体 

通过在文件 github.com/pkg/errors/errors.go 中增加新的withCode 结构体，来引入一种新的错误类型，该错误类型可以记录错误码、stack、cause 和具体的错误信息。

```go
type withCode struct {
   err   error   // error 错误
   code  int  // 业务错误码
   cause error  // cause error
   *stack  // 错误堆栈
}
```

下面，通过一个示例，来了解下github.com/marmotedu/errors 所提供的功能。 假设下述代码保存在errors.go文件中：

```go
package main

import (
	"fmt"

	"github.com/marmotedu/errors"
	code "github.com/marmotedu/sample-code"
)

func main() {
	if err := bindUser(); err != nil {
		// %s: Returns the user-safe error string mapped to the error code or the error message if none is specified.
		fmt.Println("====================> %s <====================")
		fmt.Printf("%s\n\n", err)

		// %v: Alias for %s.
		fmt.Println("====================> %v <====================")
		fmt.Printf("%v\n\n", err)

		// %-v: Output caller details, useful for troubleshooting.
		fmt.Println("====================> %-v <====================")
		fmt.Printf("%-v\n\n", err)

		// %+v: Output full error stack details, useful for debugging.
		fmt.Println("====================> %+v <====================")
		fmt.Printf("%+v\n\n", err)

		// %#-v: Output caller details, useful for troubleshooting with JSON formatted output.
		fmt.Println("====================> %#-v <====================")
		fmt.Printf("%#-v\n\n", err)

		// %#+v: Output full error stack details, useful for debugging with JSON formatted output.
		fmt.Println("====================> %#+v <====================")
		fmt.Printf("%#+v\n\n", err)

		// do some business process based on the error type
		if errors.IsCode(err, code.ErrEncodingFailed) {
			fmt.Println("this is a ErrEncodingFailed error")
		}

		if errors.IsCode(err, code.ErrDatabase) {
			fmt.Println("this is a ErrDatabase error")
		}

		// we can also find the cause error
		fmt.Println(errors.Cause(err))
	}
}

func bindUser() error {
	if err := getUser(); err != nil {
		// Step3: Wrap the error with a new error message and a new error code if needed.
		return errors.WrapC(err, code.ErrEncodingFailed, "encoding user 'Lingfei Kong' failed.")
	}

	return nil
}

func getUser() error {
	if err := queryDatabase(); err != nil {
		// Step2: Wrap the error with a new error message.
		return errors.Wrap(err, "get user failed.")
	}

	return nil
}

func queryDatabase() error {
	// Step1. Create error with specified error code.
	return errors.WithCode(code.ErrDatabase, "user 'Lingfei Kong' not found.")
}
```

上述代码中，通过 WithCode函数来创建新的 withCode 类型的错误；通过 WrapC来将一个 error 封装成一个 withCode 类型的错误；通过 IsCode来判断一个 error 链中是否包含指定的 code。 

withCode 错误实现了一个 `func (w *withCode) Format(state fmt.State, verb rune)` 方法，该方法用来打印不同格式的错误信息，见下表：

![image-20211119215915633](IAM-document.assets/image-20211119215915633.png)

例如，%+v会打印以下错误信息：

```sh
encoding user 'Lingfei Kong' failed. - #2 [/Users/rmliu/CodeStudy/MyGo/Code/src/business_errors/errors/main/errors.go:53 (main.bindUser)] (100301) Encoding failed due to an error with the data; get user failed. - #1 [/Users/rmliu/CodeStudy/MyGo/Code/src/business_errors/errors/main/errors.go:62 (main.getUser)] (100101) Database error; user 'Lingfei Kong' not found. - #0 [/Users/rmliu/CodeStudy/MyGo/Code/src/business_errors/errors/main/errors.go:70 (main.queryDatabase)] (100101) Database error
```

那么，这些错误信息中的100101错误码，还有Database error这种对外展示的报错信息等等，是从哪里获取的？

首先， withCode 中包含了 int 类型的错误码，例如100101。

其次，当使用github.com/marmotedu/errors包的时候，需要调用Register或者 MustRegister，将一个 Coder 注册到github.com/marmotedu/errors开辟的内存中，数据结构为：

```go
var codes = map[int]Coder{}
```

#### Coder 接口

Coder 是一个接口，定义为：

```go
// Coder defines an interface for an error code detail information.
type Coder interface {
	// HTTP status that should be used for the associated error code.
	HTTPStatus() int

	// External (user) facing error text.
	String() string

	// Reference returns the detail documents for user.
	Reference() string

	// Code returns the code of the coder
	Code() int
}
```

这样 withCode 的Format方法，就能够通过 withCode 中的 code 字段获取到对应的 Coder，并通过 Coder 提供的 HTTPStatus、String、Reference、Code 函数，来获取 withCode 中 code 的详细信息，最后格式化打印。 

#### Register和MustRegister 注册函数

这里要注意，实现了两个注册函数：Register和MustRegister，二者唯一区别是：当重复定义同一个错误 Code 时，MustRegister会 panic，这样可以防止后面注册的错误覆盖掉之前注册的错误。

在实际开发中，建议使用MustRegister。 

XXX()和MustXXX()的函数命名方式，是一种 Go 代码设计技巧，在 Go 代码中经常使用，例如 Go 标准库中regexp包提供的Compile和MustCompile函数。和XXX相比， MustXXX 会在某种情况不满足时 panic。

因此使用MustXXX的开发者看到函数名就会有一个心理预期：使用不当，会造成程序 panic。 

最后，还有一个建议：在实际的生产环境中，可以使用 JSON 格式打印日志，JSON 格式的日志可以非常方便的供日志系统解析。可以根据需要，选择%#-v或%#+v两种格式。 

错误包在代码中，经常被调用，所以要保证错误包一定要是高性能的，否则很可能会影响接口的性能。这里，再来看下github.com/marmotedu/errors包的性能。

#### errors 包对比

在这里，把这个错误包跟 go 标准库的 errors 包，以及 github.com/pkg/errors 包进行对比，来看看它们的性能：

```go
$ go test -test.bench=BenchmarkErrors -benchtime="3s"
goos: linux
goarch: amd64
pkg: github.com/marmotedu/errors
BenchmarkErrors/errors-stack-10-8 57658672 61.8 ns/op  16 B/op1 allocs/op         
BenchmarkErrors/pkg/errors-stack-10-8 2265558 1547 ns/op 320 B/op         
BenchmarkErrors/marmot/errors-stack-10-8 1903532 1772 ns/op 360
BenchmarkErrors/errors-stack-100-8 4883659 734 ns/op  16
BenchmarkErrors/pkg/errors-stack-100-8 1202797 2881 ns/op 320
BenchmarkErrors/marmot/errors-stack-100-8 1000000 3116 ns/op 360
BenchmarkErrors/errors-stack-1000-8 505636 7159 ns/op  16
BenchmarkErrors/pkg/errors-stack-1000-8 327681 10646 ns/op 320
BenchmarkErrors/marmot/errors-stack-1000-8 304160 11896 ns/o
PASS
ok github.com/marmotedu/errors 39.200s
```

可以看到github.com/marmotedu/errors和github.com/pkg/errors包的性能基本持平。在对比性能时，重点关注 ns/op，也即每次 error 操作耗费的纳秒数。

另外，还需要测试不同 error 嵌套深度下的 error 操作性能，嵌套越深，性能越差。例如：在嵌套深度为 10 的时候， github.com/pkg/errors 包 ns/op 值为 1547， github.com/marmotedu/errors 包 ns/op 值为 1772。可以看到，二者性能基本保持 一致。 具体性能数据对比见下表：

![image-20211119223053926](IAM-document.assets/image-20211119223053926.png)

通过 BenchmarkErrors测试函数来测试 error 包性能。

### 如何记录错误？ 

上面，一起看了怎么设计一个优秀的错误包，那如何用设计的错误包来记录错误 呢？ 

根据开发经验，推荐两种记录错误的方式，可以快速定位问题。 

#### 错误堆栈跟踪错误

方式一：通过github.com/marmotedu/errors包提供的错误堆栈能力，来跟踪错误。 

具体可以看看下面的代码示例。以下代码保存在errortrack_errors.go中。

```go
package main

import (
	"fmt"

	"github.com/marmotedu/errors"

	code "github.com/marmotedu/sample-code"
)

func main() {
	if err := getUser(); err != nil {
		fmt.Printf("%+v\n", err)
	}
}

func getUser() error {
	if err := queryDatabase(); err != nil {
		return errors.Wrap(err, "get user failed.")
	}

	return nil
}

func queryDatabase() error {
	return errors.WithCode(code.ErrDatabase, "user 'Lingfei Kong' not found.")
}
```

执行上述的代码：

```go
$ go run errortrack_errors.go
get user failed. - #1 [/Users/rmliu/CodeStudy/MyGo/Code/src/business_errors/errortrack_errors.go:19 (main.getUser)] (100101) Database error; user 'Lingfei Kong' not found. - #0 [/Users/rmliu/CodeStudy/MyGo/Code/src/business_errors/errortrack_errors.go:26 (main.queryDatabase)] (100101) Database error
```

可以看到，打印的日志中打印出了详细的错误堆栈，包括错误发生的函数、文件名、行号 和错误信息，通过这些错误堆栈，可以很方便地定位问题。 

使用这种方法时，推荐的用法是，在错误最开始处使用 errors.WithCode() 创建一个 withCode 类型的错误。上层在处理底层返回的错误时，可以根据需要，使用 Wrap 函 数基于该错误封装新的错误信息。如果要包装的 error 不是用 github.com/marmotedu/errors包创建的，建议用 errors.WithCode() 新建一个 error。 

#### 日志包记录函数

方式二：在错误产生的最原始位置调用日志包记录函数，打印错误信息，其他位置直接返回（当然，也可以选择性的追加一些错误信息，方便故障定位）。

示例代码（保存在 errortrack_log.go）如下：

```go
package main

import (
   "fmt"

   "github.com/marmotedu/errors"
   "github.com/marmotedu/log"

   code "github.com/marmotedu/sample-code"
)

func main() {
   if err := getUser(); err != nil {
      fmt.Printf("%v\n", err)
   }
}

func getUser() error {
   if err := queryDatabase(); err != nil {
      return err
   }

   return nil
}

func queryDatabase() error {
   opts := &log.Options{
      Level:            "info",
      Format:           "console",
      EnableColor:      true,
      EnableCaller:     true,
      OutputPaths:      []string{"test.log", "stdout"},
      ErrorOutputPaths: []string{},
   }

   log.Init(opts)
   defer log.Flush()

   err := errors.WithCode(code.ErrDatabase, "user 'Lingfei Kong' not found.")
   if err != nil {
      log.Errorf("%v", err)
   }
   return err
}
```

执行以上代码：

```sh
$ go run errortrack_log.go
2021-11-19 22:39:14.935 ERROR   main/errortrack_log.go:41       Database error
Database error
```

当错误发生时，调用 log 包打印错误。通过 log 包的 caller 功能，可以定位到 log 语句的位置，也就是定位到错误发生的位置。使用这种方式来打印日志时，有两个建议。

- 只在错误产生的最初位置打印日志，其他地方直接返回错误，一般不需要再对错误进行封装。 
- 当代码调用第三方包的函数时，第三方包函数出错时打印错误信息。比如：

```go
if err := os.Chdir("/root"); err != nil {
	log.Errorf("change dir failed: %v", err)
}
```

### 一个错误码的具体实现 

接下来，看一个依据上一讲介绍的错误码规范的具体错误码实现 github.com/marmotedu/sample-code。 

sample-code实现了两类错误码，分别是通用错误码（sample-code/base.go）和业务模块相关的错误码（sample-code/apiserver.go）。 

#### 通用错误码

首先，来看通用错误码的定义：

```go
/ 通用: 基本错误
// Code must start with 1xxxxx
const (
   // ErrSuccess - 200: OK.
   ErrSuccess int = iota + 100001

   // ErrUnknown - 500: Internal server error.
   ErrUnknown

   // ErrBind - 400: Error occurred while binding the request body to the struct.
   ErrBind

   // ErrValidation - 400: Validation failed.
   ErrValidation

   // ErrTokenInvalid - 401: Token invalid.
   ErrTokenInvalid
)
```

在代码中，通常使用整型常量（ErrSuccess）来代替整型错误码（100001），因为使用 ErrSuccess 时，一看就知道它代表的错误类型，可以方便开发者使用。 

错误码用来指代一个错误类型，该错误类型需要包含一些有用的信息，例如对应的 HTTP Status Code、对外展示的 Message，以及跟该错误匹配的帮助文档。

所以，还需要实现一个 Coder 来承载这些信息。这里，定义了一个实现了 github.com/marmotedu/errors.Coder接口的ErrCode结构体：

```go
// ErrCode implements `github.com/marmotedu/errors`.Coder interface.
type ErrCode struct {
   // C refers to the code of the ErrCode.
   C int

   // HTTP status that should be used for the associated error code.
   HTTP int

   // External (user) facing error text.
   Ext string

   // Ref specify the reference document.
   Ref string
}
```

可以看到ErrCode结构体包含了以下信息：

- int 类型的业务码。 
- 对应的 HTTP Status Code。 
- 暴露给外部用户的消息。 
- 错误的参考文档。

下面是一个具体的 Coder 示例：

```go
coder := &ErrCode{
  C: 100001,
  HTTP: 200,
  Ext: "OK",
  Ref: "https://github.com/marmotedu/sample-code/blob/master/README.md",
}
```

接下来，就可以调用github.com/marmotedu/errors包提供的Register或者 MustRegister函数，将 Coder 注册到github.com/marmotedu/errors包维护的内存中。 

一个项目有很多个错误码，如果每个错误码都手动调用MustRegister函数会很麻烦，这里通过代码自动生成的方法，来生成 register 函数调用：

```go
//go:generate codegen -type=int
//go:generate codegen -type=int -doc -output ./error_code_generated.md
```

`//go:generate codegen -type=int` 会调用 codegen工具，生成  sample_code_generated.go源码文件：

```go
func init() {
  register(ErrSuccess, 200, "OK")
  register(ErrUnknown, 500, "Internal server error")
  register(ErrBind, 400, "Error occurred while binding the request body to the
  register(ErrValidation, 400, "Validation failed")
  // other register function call
}
```

这些 register调用放在 init 函数中，在加载程序的时候被初始化。

这里要注意，在注册的时候，会检查 HTTP Status Code，只允许定义 200、400、 401、403、404、500 这 6 个 HTTP 错误码。这里通过程序保证了错误码是符合 HTTP Status Code 使用要求的。 

`//go:generate codegen -type=int -doc -output ./error_code_generated.md` 会生成错误码描述文档 error_code_generated.md。 当我们提供 API 文档时，也需要记着提供一份错误码描述文档，这样客户端才可以根据错 误码，知道请求是否成功，以及具体发生哪类错误，好针对性地做一些逻辑处理。 

codegen工具会根据错误码注释生成sample_code_generated.go和 error_code_generated.md文件：

```go
// ErrSuccess - 200: OK.
ErrSuccess int = iota + 100001
```

codegen 工具之所以能够生成sample_code_generated.go和 error_code_generated.md，是因为错误码注释是有规定格式的：// <错误码 整型常量> - <对应的HTTP Status Code>: <External Message>.。 

codegen 工具可以在 IAM 项目根目录下，执行以下命令来安装：

```sh
$ make tools.install.codegen
```

安装完 codegen 工具后，可以在 github.com/marmotedu/sample-code 包根目录下执行 go generate 命令，来生成sample_code_generated.go和 error_code_generated.md。

这里有个技巧需要注意：生成的文件建议统一用 xxxx_generated.xx 来命名，这样通过 generated ，就知道这个文件是代码自动生成的，有助于理解和使用。

在实际的开发中，可以将错误码独立成一个包，放在 internal/pkg/code/目录下，这样可以方便整个应用调用。

例如 IAM 的错误码就放在 IAM 项目根目录下的internal/pkg/code/目录下。

错误码是分服务和模块的，所以这里建议把相同的服务放在同一个 Go 源文件中，例如 IAM 的错误码存放文件：

```sh
$ ls base.go apiserver.go authzserver.go
apiserver.go authzserver.go base.go
```

#### 业务模块相关的错误码

一个应用中会有多个服务，例如 IAM 应用中，就包含了 iam-apiserver、iam-authzserver、iam-pump 三个服务。这些服务有一些通用的错误码，为了便于维护，可以将这些通用的错误码统一放在 base.go 源码文件中。

其他的错误码，可以按服务分别放在不同的文件中：iam-apiserver 服务的错误码统一放在 apiserver.go 文件中；iam-authzserver 的错误码统一存放在 authzserver.go 文件中。其他服务以此类推。 

另外，同一个服务中不同模块的错误码，可以按以下格式来组织：相同模块的错误码放在同一个 const 代码块中，不同模块的错误码放在不同的 const 代码块中。每个 const 代码块的开头注释就是该模块的错误码定义。例如：

```go
// iam-apiserver: user errors.
const (
   // ErrUserNotFound - 404: User not found.
   ErrUserNotFound int = iota + 110001
   // ErrUserAlreadyExist - 400: User already exist.
   ErrUserAlreadyExist
)

// iam-apiserver: secret errors.
const (
   // ErrEncrypt - 400: Secret reach the max count.
   ErrReachMaxCount int = iota + 110101
   // ErrSecretNotFound - 404: Secret not found.
   ErrSecretNotFound
)
```

最后，还需要将错误码定义记录在项目的文件中，供开发者查阅、遵守和使用，例如 IAM 项目的错误码定义记录文档为code_specification.md。

这个文档中记录了错误码说明、错误描述规范和错误记录规范等。 

### 错误码实际使用方法示例 

上面，讲解了错误包和错误码的实现方式，在实际开发中是如何使用的。这里，举一个在 gin web 框架中使用该错误码的例子：

```go
// Response defines project response format which in marmotedu organization.
type Response struct {
   Code      errors.Code `json:"code,omitempty"`
   Message   string      `json:"message,omitempty"`
   Reference string      `json:"reference,omitempty"`
   Data      interface{} `json:"data,omitempty"`
}

// WriteResponse used to write an error and JSON data into response.
func WriteResponse(c *gin.Context, err error, data interface{}) {
   if err != nil {
      coder := errors.ParseCoder(err)
      c.JSON(coder.HTTPStatus(), Response{
         Code:      coder.Code(),
         Message:   coder.String(),
         Reference: coder.Reference(),
         Data:      data,
      })
   }
   c.JSON(http.StatusOK, Response{Data: data})
}
func GetUser(c *gin.Context) {
   log.Info("get user function called.", "X-Request-Id", requestid.Get(c))
   // Get the user by the `username` from the database.
   user, err := store.Client().Users().Get(c.Param("username"), metav1.GetOptions{})
   if err != nil {
      core.WriteResponse(c, code.ErrUserNotFound.Error(), nil)
      return
   }
   core.WriteResponse(c, nil, user)
}
```

上述代码中，通过WriteResponse统一处理错误。

在 WriteResponse 函数中，如果 err != nil，则从 error 中解析出 Coder，并调用 Coder 提供的方法，获取错误相关的 Http Status Code、int 类型的业务码、暴露给用户的信息、错误的参考文档链接，并返回 JSON 格式的信息。

如果 err == nil 则返回 200 和数据。 

### 总结 

记录错误是应用程序必须要做的一件事情，在实际开发中，通常会封装自己的错误包。一个优秀的错误包，应该能够支持错误堆栈、不同的打印格式、Wrap/Unwrap/Is/As 等函数，并能够支持格式化创建 error。 

根据这些错误包设计要点，基于 github.com/pkg/errors 包设计了 IAM 项目的错误包 github.com/marmotedu/errors ，该包符合上一讲设计的错误码规范。 

另外，也给出了一个具体的错误码实现 sample-code ， sample-code 支持业务 Code 码、HTTP Status Code、错误参考文档、可以对内对外展示不同的错误信息。 

最后，因为错误码注释是有固定格式的，所以可以通过 codegen 工具解析错误码的注释，生成 register 函数调用和错误码文档。这种做法也体现了一直强调的 low code 思 想，可以提高开发效率，减少人为失误。 

### 课后练习

- 在这门课里，定义了 base、iam-apiserver 服务的错误码，请试着定义 iamauthz-server 服务的错误码，并生成错误码文档。 
- 思考下，这门课的错误包和错误码设计能否满足当前的项目需求。



## Go 项目之设计日志包

来聊聊如何设计和开发日志包。 

在做 Go 项目开发时，除了处理错误之外，必须要做的另外一件事是记录日志。通过记录日志，可以完成一些基本功能，比如开发、测试期间的 Debug，故障排除，数据分 析，监控告警，以及记录发生的事件等。 

要实现这些功能，首先需要一个优秀的日志包。另外，还发现不少 Go 项目开发者 记录日志很随意，输出的日志并不能有效定位到问题。所以，还需要知道怎么更好地 记录日志，这就需要一个日志记录规范。 

有了优秀的日志包和日志记录规范，就能很快地定位到问题，获取足够的信息，并完 成后期的数据分析和监控告警，也可以很方便地进行调试了。这一讲，就来详细介绍下，如何设计日志包和日志记录规范。 

首先，来看下如何设计日志包。 

### 如何设计日志包 

目前，虽然有很多优秀的开源日志包可供选择，但在一个大型系统中，这些开源日志 包很可能无法满足定制化的需求，这时候就需要自己开发日志包。 

这些日志包可能是基于某个，或某几个开源的日志包改造而来，也可能是全新开发的日志 包。那么在开发日志包时，需要实现哪些功能，又如何实现呢？

先来看下日志包需要具备哪些功能。根据功能的重要性，将日志包需要实现的功能分为 基础功能、高级功能和可选功能。

- 基础功能是一个日志包必须要具备的功能；
- 高级功能、 可选功能都是在特定场景下可增加的功能。

### 基础功能 

基础功能，是优秀日志包必备的功能，能够满足绝大部分的使用场景，适合一些中小型的 项目。一个日志包应该具备以下 4 个基础功能。

#### 基本的日志信息

1. 支持基本的日志信息

日志包需要支持基本的日志信息，包括时间戳、文件名、行号、日志级别和日志信息。 

**时间戳**可以记录日志发生的时间。在定位问题时，往往需要根据时间戳，来复原请求 过程，核对相同时间戳下的上下文，从而定位出问题。 

**文件名和行号**，可以更快速定位到打印日志的位置，找到问题代码。一个日志库如 果不支持文件名和行号，排查故障就会变得非常困难，基本只能靠 grep 和记忆来定位代 码。对于企业级的服务，需要保证服务在故障后能够快速恢复，恢复的时间越久，造成的 损失就越大，影响就越大。这就要求研发人员能够快速定位并解决问题。通过文件名和行 号，可以精准定位到问题代码，尽快地修复问题并恢复服务。

通过**日志级别**，可以知道日志的错误类型，最通常的用法是：直接过滤出 Error 级别的日 志，这样就可以直接定位出问题出错点，然后再结合其他日志定位出出错的原因。如果不 支持日志级别，在定位问题时，可能要查看一大堆无用的日志。在大型系统中，一次请求 的日志量很多，这会大大延长定位问题的时间。 

而通过**日志信息**，可以知道错误发生的具体原因。

#### 不同的日志级别

2. 支持不同的日志级别

不同的日志级别代表不同的日志类型，例如：

- Error 级别的日志，说明日志是错误类型，在 排障时，会首先查看错误级别的日志。
- Warn 级别日志说明出现异常，但还不至于影响程 序运行，如果程序执行的结果不符合预期，则可以参考 Warn 级别的日志，定位出异常所 在。
- Info 级别的日志，可以协助 Debug，并记录一些有用的信息，供后期进行分析。 

通常一个日志包至少要实现 6 个级别，提供了一张表格，按优先级从低到高排列如 下：

![image-20211123002117136](IAM-document.assets/image-20211123002117136.png)

有些日志包，例如 logrus，还支持 Trace 日志级别。Trace 级别比 Debug 级别还低，能 够打印更细粒度的日志信息。Trace 级别不是必须的，可以根据需要自行选择。 

打印日志时，一个日志调用其实具有两个属性：

- 输出级别：打印日志时，期望日志的输出级别。例如，调用 glog.Info("This is info message") 打印一条日志，则输出日志级别为 Info。 
- 开关级别：启动应用程序时，期望哪些输出级别的日志被打印。例如，使用 glog 时 - v=4 ，说明了只有日志级别高于 4 的日志才会被打印。

如果开关级别设置为 L ，只有输出级别 >=L 时，日志才会被打印。例如，开关级别为 Warn，则只会记录 Warn、Error 、Panic 和 Fatal 级别的日志。具体的输出关系如下图所 示：

![image-20211123002335706](IAM-document.assets/image-20211123002335706.png)

#### 自定义配置

3. 支持自定义配置

不同的运行环境，需要不同的日志输出配置，例如：

- 开发测试环境为了能够方便地 Debug，需要设置日志级别为 Debug 级别；
- 现网环境为了提高应用程序的性能，则需要 设置日志级别为 Info 级别。
- 又比如，现网环境为了方便日志采集，通常会输出 JSON 格式 的日志；
- 开发测试环境为了方便查看日志，会输出 TEXT 格式的日志。 

所以，日志包需要能够被配置，还要不同环境采用不同的配置。通过配置，可以在 不重新编译代码的情况下，改变记录日志的行为。

#### 输出到标准输出和文件

4. 支持输出到标准输出和文件

日志总是要被读的，要么输出到标准输出，供开发者实时读取，要么保存到文件，供开发 者日后查看。输出到标准输出和保存到文件是一个日志包最基本的功能。 

### 高级功能 

除了上面提到的这些基本功能外，在一些大型系统中，通常还会要求日志包具备一些高级 功能。这些高级功能可以更好地记录日志，并实现更丰富的功能，例如日志告警。 那么一个日志包可以具备哪些高级功能呢？

#### 多种日志格式

1. 支持多种日志格式

日志格式也是要考虑的一个点，一个好的日志格式，不仅方便查看日志，还能方便一 些日志采集组件采集日志，并对接类似 Elasticsearch 这样的日志搜索引擎。 

一个日志包至少需要提供以下两种格式：

- TEXT 格式：TEXT 格式的日志具有良好的可读性，可以方便在开发联调阶段查看日 志，例如：

  - ```go
    2020-12-02T01:16:18+08:00 INFO example.go:11 std log
    2020-12-02T01:16:18+08:00 DEBUG example.go:13 change std log to debug level
    ```

- JSON 格式：JSON 格式的日志可以记录更详细的信息，日志中包含一些通用的或自定 义的字段，可供日后的查询、分析使用，而且可以很方便地供 filebeat、logstash 这类 日志采集工具采集并上报。下面是 JSON 格式的日志：

  - ```go
    {"level":"DEBUG","time":"2020-12-02T01:16:18+08:00","file":"example.go:15","func":"main.main","message":"log in json format"}
    {"level":"INFO","time":"2020-12-02T01:16:18+08:00","file":"example.go:16","func":"main.main","message":"another log in json format"}
    ```

建议在开发联调阶段使用 TEXT 格式的日志，在现网环境使用 JSON 格式的日志。一个 优秀的日志库，例如 logrus，除了提供基本的输出格式外，还应该允许开发者自定义日志 输出格式。

#### 按级别分类输出

2. 能够按级别分类输出

为了能够快速定位到需要的日志，一个比较好的做法是将日志按级别分类输出，至少错误 级别的日志可以输出到独立的文件中。这样，出现问题时，可以直接查找错误文件定位问 题。

例如，glog 就支持分类输出，如下图所示：

![image-20211123003558928](IAM-document.assets/image-20211123003558928.png)

#### 结构化日志

3. 支持结构化日志

结构化日志（Structured Logging），就是使用 JSON 或者其他编码方式使日志结构化， 这样可以方便后续使用 Filebeat、Logstash Shipper 等各种工具，对日志进行采集、过 滤、分析和查找。就像下面这个案例，使用 zap 进行日志打印：

```go
package main

import (
   "go.uber.org/zap"
   "time"
)

func main() {
   logger, _ := zap.NewProduction()
   defer logger.Sync() // flushes buffer, if any
   url := "http://marmotedu.com"
   // 结构化日志打印
   logger.Sugar().Infow("failed to fetch URL", "url", url, "attempt", 3, "backoff", time.Second)
   // 非结构化日志打印
   logger.Sugar().Infof("failed to fetch URL: %s", url)
}
```

上述代码输出为：

```go
{"level":"info","ts":1637599210.365768,"caller":"main/zap_example.go:13","msg":"failed to fetch URL","url":"http://marmotedu.com","attempt":3,"backoff":1}
{"level":"info","ts":1637599210.365906,"caller":"main/zap_example.go:15","msg":"failed to fetch URL: http://marmotedu.com"}
```

#### 日志轮转

4. 支持日志轮转

在一个大型项目中，一天可能会产生几十个 G 的日志。为了防止日志把磁盘空间占满，导 致服务器或者程序异常，就需要确保日志大小达到一定量级时，对日志进行切割、压缩， 并转存。 

如何切割呢？可以按照日志大小进行切割，也可以按日期切割。日志的切割、压缩和转 存功能可以基于 GitHub 上一些优秀的开源包来封装，例如：

- lumberjack可以支持按大 小和日期归档日志，
- file-rotatelogs支持按小时数进行日志切割。 

对于日志轮转功能，其实不建议在日志包中添加，因为这会增加日志包的复杂度，建议的做法是借助其他的工具来实现日志轮转。例如，在 Linux 系统中可以使用 Logrotate 来轮转日志。Logrotate 功能强大，是一个专业的日志轮转工具。

#### Hook 能力

5. 具备 Hook 能力

Hook 能力可以对日志内容进行自定义处理。例如，当某个级别的日志产生时，发 送邮件或者调用告警接口进行告警。很多优秀的开源日志包提供了 Hook 能力，例如 logrus 和 zap。 

在一个大型系统中，日志告警是非常重要的功能，但更好的实现方式是将告警能力做成旁 路功能。通过旁路功能，可以保证日志包功能聚焦、简洁。例如：可以将日志收集到 Elasticsearch，并通过 ElastAlert 进行日志告警

### 可选功能 

除了基础功能和高级功能外，还有一些功能。这些功能不会影响到日志包的核心功能，但 是如果具有这些功能，会使日志包更加易用。比如下面的这三个功能。

#### 颜色输出

1. 支持颜色输出

在开发、测试时开启颜色输出，不同级别的日志会被不同颜色标识，这样可以很轻松 地发现一些 Error、Warn 级别的日志，方便开发调试。发布到生产环境时，可以关闭颜色 输出，以提高性能。

#### 兼容标准库log包

2. 兼容标准库 log 包

一些早期的 Go 项目大量使用了标准库 log 包，如果日志库能够兼容标准库 log 包，就可以很容易地替换掉标准库 log 包。例如，logrus 就兼容标准库 log 包。这 里，来看一个使用了标准库 log 包的代码：

```go
import (
	"log"
)

func main() {
	log.Print("call Print: line1")
	log.Println("call Println: line2")
}
```

只需要使用log "github.com/sirupsen/logrus"替换"log"就可以完成标准库 log 包的切换：

```go
import (
   log "github.com/sirupsen/logrus"
)

func main() {
   log.Print("call Print: line1")
   log.Println("call Println: line2")
}
```

#### 输出到不同的位置

3. 支持输出到不同的位置

在分布式系统中，一个服务会被部署在多台机器上，这时候如果要查看日志，就需要 分别登录不同的机器查看，非常麻烦。更希望将日志统一投递到 Elasticsearch 上，在 Elasticsearch 上查看日志。 

还可能需要从日志中分析某个接口的调用次数、某个用户的请求次数等信息，这就需要能够对日志进行处理。一般的做法是将日志投递到 Kafka，数据处理服务消费 Kafka 中保存的日志，从而分析出调用次数等信息。 

以上两种场景，分别需要把日志投递到 Elasticsearch、Kafka 等组件，如果日志包 支持将日志投递到不同的目的端，那会是一项非常让人期待的功能：

![image-20211123005919106](IAM-document.assets/image-20211123005919106.png)

如果日志不支持投递到不同的下游组件，例如 Elasticsearch、Kafka、Fluentd、 Logstash 等位置，也可以通过 Filebeat 采集磁盘上的日志文件，进而投递到下游组件。 

### 设计日志包时需要关注的点 

上面，介绍了日志包具备的功能，这些功能可以指导完成日志包设计。这里，再来看下设计日志包时，还需要关注的几个层面：

- 高性能：因为要在代码中频繁调用日志包，记录日志，所以日志包的性能是首先要 考虑的点，一个性能很差的日志包必然会导致整个应用性能很差。 
- 并发安全：Go 应用程序会大量使用 Go 语言的并发特性，也就意味着需要并发地记录日 志，这就需要日志包是并发安全的。
- 插件化能力：日志包应该能提供一些插件化的能力，比如允许开发者自定义输出格式， 自定义存储位置，自定义错误发生时的行为（例如 告警、发邮件等）。插件化的能力不 是必需的，因为日志自身的特性就能满足绝大部分的使用需求，例如：输出格式支持 JSON 和 TEXT，存储位置支持标准输出和文件，日志监控可以通过一些旁路系统来实 现。 
- 日志参数控制：日志包应该能够灵活地进行配置，初始化时配置或者程序运行时配置。 例如：初始化配置可以通过 Init 函数完成，运行时配置可以通过 SetOptions / SetLevel 等函数来完成。

### 如何记录日志？ 

前面介绍了在设计日志包时，要包含的一些功能、实现方法和注意事项。但在这个过 程中，还有一项重要工作需要注意，那就是日志记录问题。 

日志并不是越多越好，在实际开发中，经常会遇到一大堆无用的日志，却没有需要的 日志；或者有效的日志被大量无用的日志淹没，查找起来非常困难。 一个优秀的日志包可以协助更好地记录、查看和分析日志，但是如何记录日志决定了能否获取到有用的信息。日志包是工具，日志记录才是灵魂。这里，就来详细讲讲 如何记录日志。 想要更好地记录日志，需要解决以下几个问题：

- 在何处打印日志？ 
- 在哪个日志级别打印日志？ 
- 如何记录日志内容？

#### 在何处打印日志？ 

日志主要是用来定位问题的，所以整体来说，要在有需要的地方打印日志。那么具体 是哪些地方呢？给几个建议。

- 在分支语句处打印日志。在分支语句处打印日志，可以判断出代码走了哪个分支，有助 于判断请求的下一跳，继而继续排查问题。
- 写操作必须打印日志。写操作最可能会引起比较严重的业务故障，写操作打印日志，可 以在出问题时找到关键信息。
- 在循环中打印日志要慎重。如果循环次数过多，会导致打印大量的日志，严重拖累代码 的性能，建议的办法是在循环中记录要点，在循环外面总结打印出来。 
- 在错误产生的最原始位置打印日志。对于嵌套的 Error，可在 Error 产生的最初位置打印 Error 日志，上层如果不需要添加必要的信息，可以直接返回下层的 Error。

举个例子：

```go
package main

import (
   "flag"
   "fmt"
   "github.com/golang/glog"
)

func main() {
   flag.Parse()
   defer glog.Flush()
   if err := loadConfig(); err != nil {
      glog.Error(err)
   }
}

func loadConfig() error {
   return decodeConfig() // 直接返回
}

func decodeConfig() error {
   if err := readConfig(); err != nil {
      return fmt.Errorf("could not decode configuration data for user %s: %v", "colin", err) // 添加必要的信息，用户名称
   }
   return nil
}

func readConfig() error {
   glog.Errorf("read: end of input.")
   return fmt.Errorf("read: end of input")
}
```

通过在最初产生错误的位置打印日志，可以很方便地追踪到日志的根源，进而在上层 追加一些必要的信息。这可以了解到该错误产生的影响，有助于排障。另外，直接 返回下层日志，还可以减少重复的日志打印。 

当代码调用第三方包的函数，且第三方包函数出错时，会打印错误信息。比如：

```go
if err := os.Chdir("/root"); err != nil {
  log.Errorf("change dir failed: %v", err)
}
```

#### 在哪个日志级别打印日志？ 

不同级别的日志，具有不同的意义，能实现不同的功能，在开发中，应该根据目的， 在合适的级别记录日志，这里给一些建议。

##### Debug 级别

1. Debug 级别

为了获取足够的信息进行 Debug，通常会在 Debug 级别打印很多日志。例如，可以打印 整个 HTTP 请求的请求 Body 或者响应 Body。 

Debug 级别需要打印大量的日志，这会严重拖累程序的性能。并且，Debug 级别的日 志，主要是为了能在开发测试阶段更好地 Debug，多是一些不影响现网业务的日志信息。 所以，对于 Debug 级别的日志，在服务上线时一定要禁止掉。否则，就可能会因为大 量的日志导致硬盘空间快速用完，从而造成服务宕机，也可能会影响服务的性能和产品体 验。 

Debug 这个级别的日志可以随意输出，任何觉得有助于开发、测试阶段调试的日志，都 可以在这个级别打印。

##### Info 级别

2. Info 级别

Info 级别的日志可以记录一些有用的信息，供以后的运营分析，所以 Info 级别的日志不是 越多越好，也不是越少越好，应以满足需求为主要目标。一些关键日志，可以在 Info 级别 记录，但如果日志量大、输出频度过高，则要考虑在 Debug 级别记录。

现网的日志级别一般是 Info 级别，为了不使日志文件占满整个磁盘空间，在记录日志时， 要注意避免产生过多的 Info 级别的日志。例如，在 for 循环中，就要慎用 Info 级别的日 志。

##### Warn 级别

3. Warn 级别

一些警告类的日志可以记录在 Warn 级别，Warn 级别的日志往往说明程序运行异常，不 符合预期，但又不影响程序的继续运行，或者是暂时影响，但后续会恢复。像这些日志， 就需要关注起来。Warn 更多的是业务级别的警告日志。

##### Error 级别

4. Error 级别

Error 级别的日志告诉程序执行出错，这些错误肯定会影响到程序的执行结果，例如请 求失败、创建资源失败等。要记录每一个发生错误的日志，避免日后排障过程中这些错误 被忽略掉。大部分的错误可以归在 Error 级别。

##### Panic 级别

5. Panic 级别

Panic 级别的日志在实际开发中很少用，通常只在需要错误堆栈，或者不想因为发生严重 错误导致程序退出，而采用 defer 处理错误时使用。

##### Fatal 级别

6. Fatal 级别

Fatal 是最高级别的日志，这个级别的日志说明问题已经相当严重，严重到程序无法继续运 行，通常是系统级的错误。在开发中也很少使用，除非觉得某个错误发生时，整个程 序无法继续运行。 

##### 总结

这里用一张图来总结下，如何选择 Debug、Info、Warn、Error、Panic、Fatal 这几种日 志级别。

![image-20211124000220835](IAM-document.assets/image-20211124000220835.png)

#### 如何记录日志内容？ 

关于如何记录日志内容，有几条建议：

- 在记录日志时，不要输出一些敏感信息，例如密码、密钥等。 
- 为了方便调试，通常会在 Debug 基本记录一些临时日志，这些日志内容可以用一些特 殊的字符开头，例如 `log.Debugf("XXXXXXXXXXXX-1:Input key was: %s", setKeyName) `。这样，在完成调试后，可以通过查找 XXXXXXXXXXXX 字符串，找到 这些临时日志，在 commit 前删除。 
- 日志内容应该小写字母开头，以英文点号 . 结尾，例如 log.Info("update user function called.") 。 
- 为了提高性能，尽可能使用明确的类型，例如使用 `log.Warnf("init datastore: %s", err.Error())` 而非 `log.Warnf("init datastore: %v", err)` 。 
- 根据需要，日志最好包含两个信息。
  - 一个是请求 ID（RequestID），是每次请求的唯一 ID，便于从海量日志中过滤出某次请求的日志，可以将请求 ID 放在请求的通用日志字 段中。
  - 另一个是用户和行为，用于标识谁做了什么。 
- 不要将日志记录在错误的日志级别上。例如，在项目开发中，经常会发现有同事将正 常的日志信息打印在 Error 级别，将错误的日志信息打印在 Info 级别。

#### 记录日志的“最佳”实践总结 

关于日志记录问题，从以上三个层面给讲解了。

综合来说，对于日志记录的最佳实 践，在平时都可以注意或进行尝试，把这些重点放在这里，方便后续查阅。

- 开发调试、现网故障排障时，不要遗忘一件事情：根据排障的过程优化日志打印。好的 日志，可能不是一次就可以写好的，可以在实际开发测试，还有现网定位问题时，不断 优化。但这需要重视日志，而不是把日志仅仅当成记录信息的一种方式，甚至不知道 为什么打印一条 Info 日志。 
- 打印日志要“不多不少”，避免打印没有作用的日志，也不要遗漏关键的日志信息。最 好的信息是，仅凭借这些关键的日志就能定位到问题。 
- 支持动态日志输出，方便线上问题定位。 
- 总是将日志记录在本地文件：通过将日志记录在本地文件，可以和日志中心化平台进行 解耦，这样当网络不可用，或者日志中心化平台故障时，仍然能够正常的记录日志。 
- 集中化日志存储处理：因为应用可能包含多个服务，一个服务包含多个实例，为了查看 日志方便，最好将这些日志统一存储在同一个日志平台上，例如 Elasticsearch，方便集 中管理和查看日志。 
- 结构化日志记录：添加一些默认通用的字段到每行日志，方便日志查询和分析。 
- 支持 RequestID：使用 RequestID 串联一次请求的所有日志，这些日志可能分布在不 同的组件，不同的机器上。支持 RequestID 可以大大提高排障的效率，降低排障难度。 在一些大型分布式系统中，没有 RequestID 排障简直就是灾难。 
- 支持动态开关 Debug 日志：对于定位一些隐藏得比较深的问题，可能需要更多的信 息，这时候可能需要打印 Debug 日志。但现网的日志级别会设置为 Info 级别，为了获 取 Debug 日志，可能会修改日志级别为 Debug 级别并重启服务，定位完问题后， 再修改日志级别为 Info 级别，然后再重启服务，这种方式不仅麻烦而且还可能会对现网 业务造成影响，最好的办法是能够在请求中通过 debug=true 这类参数动态控制某次请 求是否开启 Debug 日志。

### 拓展内容：分布式日志解决方案（EFK/ELK） 

前面介绍了设计日志包和记录日志的规范，除此之外，还有一个问题应该了解，那 就是：记录的日志如何收集、处理和展示。 

在实际 Go 项目开发中，为了实现高可用，同一个服务至少需要部署两个实例，通过轮询 的负载均衡策略转发请求。另外，一个应用又可能包含多个服务。假设应用包含两 个服务，每个服务部署两个实例，如果应用出故障，可能需要登陆 4（2 x 2）台服务器查看本地的日志文件，定位问题，非常麻烦，会增加故障恢复时间。

所以在真实的企业 场景中，会将这些日志统一收集并展示。 

在业界，日志的收集、处理和展示，早已经有了一套十分流行的日志解决方案： EFK（Elasticsearch + Filebeat + Kibana）或者 ELK（Elasticsearch + Logstash + Kibana），EFK 可以理解为 ELK 的演进版，把日志收集组件从 Logstash 替换成了 Filebeat。

用 Filebeat 替换 Logstash，主要原因是 Filebeat 更轻量级，占用的资源更 少。关于日志处理架构，可以参考这张图。

![image-20211124001359901](IAM-document.assets/image-20211124001359901.png)

通过 log 包将日志记录在本地文件中（*.log 文件），再通过 Shipper 收集到 Kafka 中。 Shipper 可以根据需要灵活选择，常见的 Shipper 有 Logstash Shipper、Flume、 Fluentd、Filebeat。

其中 Filebeat 和 Logstash Shipper 用得最多。Shipper 没有直接将 日志投递到 Logstash indexer，或者 Elasticsearch，是因为 Kafka 能够支持更大的吞吐 量，起到削峰填谷的作用。 

Kafka 中的日志消息会被 Logstash indexer 消费，处理后投递到 Elasticsearch 中存储起 来。

Elasticsearch 是实时全文搜索和分析引擎，提供搜集、分析、存储数据三大功能。 Elasticsearch 中存储的日志，可以通过 Kibana 提供的图形界面来展示。

Kibana 是一个基 于 Web 的图形界面，用于搜索、分析和可视化存储在 Elasticsearch 中的日志数据。 

Logstash 负责采集、转换和过滤日志。它支持几乎任何类型的日志，包括系统日志、错误 日志和自定义应用程序日志。Logstash 又分为 Logstash Shipper 和 Logstash indexer。 其中，Logstash Shipper 监控并收集日志，并将日志内容发送到 Logstash indexer，然 后 Logstash indexer 过滤日志，并将日志提交给 Elasticsearch。 

### 总结 

记录日志，是应用程序必备的功能。记录日志最大的作用是排障，如果想更好地排障，需要一个优秀的工具，日志包。那么如何设计日志包呢？首先需要知道日志包的功 能，日志包需要具备以下功能：

- 基础功能：支持基本的日志信息、支持不同的日志级别、支持自定义配置、支持输出到 标准输出和文件。 
- 高级功能：支持多种日志格式、能够按级别分类输出、支持结构化日志、支持日志轮 转、具备 Hook 能力。 
- 可选功能：支持颜色输出、兼容标准库 log 包、支持输出到不同的位置。

另外，一个日志包还需要支持不同级别的日志，按日志级别优先级从低到高分别是：Trace < Debug < Info < Warn/Warning < Error < Panic < Fatal。其中 Debug、Info、 Warn、Error、Fatal 是比较基础的级别，建议在开发一个日志包时包含这些级别。 Trace、Panic 是可选的级别。 

在掌握了日志包的功能之后，就可以设计、开发日志包了。但在开发过程中，还 需要确保日志包具有比较高的性能、并发安全、支持插件化的能力，并支持日志参 数控制。 

有了日志包，还需要知道如何更好地使用日志包，也就是如何记录日志。在文中，给出了一些记录建议，内容比较多，可以返回文中查看。 

最后，还给出了分布式日志解决方案：EFK/ELK。EFK 是 ELK 的升级版，在实际项目开 发中，可以直接选择 EFK。在 EFK 方案中，通过 Filebeat 将日志上传到 Kafka， Logstash indexer 消费 Kafka 中的日志，并投递到 Elasticsearch 中存储起来，最后通过 Kibana 图形界面来查看日志。 

### 课后练习

思考一下，项目中，日志包还需要哪些功能，如何设计？日常开发中，如果有比 较好的日志记录规范，也欢迎分享讨论。



## Go 项目之编写日志包

上一讲介绍了如何设计日志包，今天是实战环节，会手把手从 0 编写一个日志 包。 

在实际开发中，可以选择一些优秀的开源日志包，不加修改直接拿来使用。但更多时 候，是基于一个或某几个优秀的开源日志包进行二次开发。想要开发或者二次开发一个日 志包，就要掌握日志包的实现方式。

那么这一讲中，来从 0 到 1，实现一个具备基本功能的日志包，从中一窥日志包的实现原理和实现方法。 

在开始实战之前，先来看下目前业界有哪些优秀的开源日志包。

### 有哪些优秀的开源日志包？ 

在 Go 项目开发中，可以通过修改一些优秀的开源日志包，来实现项目的日志包。Go 生态中有很多优秀的开源日志包，例如标准库 log 包、glog、logrus、zap、seelog、 zerolog、log15、apex/log、go-logging 等。

其中，用得比较多的是标准库 log 包、 glog、logrus 和 zap。 

为了解开源日志包的现状，接下来会简单介绍下这几个常用的日志包。至于它们 的具体使用方法，可以参考整理的一篇文章：优秀开源日志包使用教程。 

#### 标准库 log 包 

标准库 log 包的功能非常简单，只提供了 Print、Panic 和 Fatal 三类函数用于日志输出。 因为是标准库自带的，所以不需要下载安装，使用起来非常方便。 

标准库 log 包只有不到 400 行的代码量，如果想研究如何实现一个日志包，阅读标准库 log 包是一个不错的开始。Go 的标准库大量使用了 log 包，例如net/http 、 net/rpc 等。 

#### glog 

glog是 Google 推出的日志包，跟标准库 log 包一样，它是一个轻量级的日志包，使用 起来简单方便。但 glog 比标准库 log 包提供了更多的功能，它具有如下特性：

- 支持 4 种日志级别：Info、Warning、Error、Fatal。 
- 支持命令行选项，例如-alsologtostderr、-log_backtrace_at、-log_dir、- logtostderr、-v等，每个参数实现某种功能。 
- 支持根据文件大小切割日志文件。 
- 支持日志按级别分类输出。 
- 支持 V level。V level 特性可以使开发者自定义日志级别。 
- 支持 vmodule。vmodule 可以使开发者对不同的文件使用不同的日志级别。 
- 支持 traceLocation。traceLocation 可以打印出指定位置的栈信息。

Kubernetes 项目就使用了基于 glog 封装的 klog，作为其日志库。

#### logrus 

logrus是目前 GitHub 上 star 数量最多的日志包，它的优点是功能强大、性能高效、高 度灵活，还提供了自定义插件的功能。很多优秀的开源项目，例如 Docker、Prometheus 等，都使用了 logrus。除了具有日志的基本功能外，logrus 还具有如下特性：

- 支持常用的日志级别。logrus 支持 Debug、Info、Warn、Error、Fatal 和 Panic 这些 日志级别。 
- 可扩展。logrus 的 Hook 机制允许使用者通过 Hook 的方式，将日志分发到任意地 方，例如本地文件、标准输出、Elasticsearch、Logstash、Kafka 等。 
- 支持自定义日志格式。logrus 内置了 JSONFormatter 和 TextFormatter 两种格式。除此之外，logrus 还允许使用者通过实现 Formatter 接口，来自定义日志格式。 
- 结构化日志记录。logrus 的 Field 机制允许使用者自定义日志字段，而不是通过冗长的 消息来记录日志。 
- 预设日志字段。logrus 的 Default Fields 机制，可以给一部分或者全部日志统一添加共 同的日志字段，例如给某次 HTTP 请求的所有日志添加 X-Request-ID 字段。 
- Fatal handlers。logrus 允许注册一个或多个 handler，当产生 Fatal 级别的日志时调 用。当程序需要优雅关闭时，这个特性会非常有用。

#### zap 

zap是 uber 开源的日志包，以高性能著称，很多公司的日志包都是基于 zap 改造而来。 除了具有日志基本的功能之外，zap 还具有很多强大的特性：

- 支持常用的日志级别，例如：Debug、Info、Warn、Error、DPanic、Panic、Fatal。 
- 性能非常高。zap 具有非常高的性能，适合对性能要求比较高的场景。 
- 支持针对特定的日志级别，输出调用堆栈。 
- 像 logrus 一样，zap 也支持结构化的目录日志、预设日志字段，也因为支持 Hook 而 具有可扩展性。

### 开源日志包选择

上面介绍了很多日志包，每种日志包使用的场景不同，可以根据自己的需求，结合日 志包的特性进行选择：

- 标准库 log 包： 标准库 log 包不支持日志级别、日志分割、日志格式等功能，所以在大 型项目中很少直接使用，通常用于一些短小的程序，比如用于生成 JWT Token 的 main.go 文件中。标准库日志包也很适合一些简短的代码，用于快速调试和验证。 
- glog： glog 实现了日志包的基本功能，非常适合一些对日志功能要求不多的小型项 目。 
- logrus： logrus 功能强大，不仅实现了日志包的基本功能，还有很多高级特性，适合 一些大型项目，尤其是需要结构化日志记录的项目。 
- zap： zap 提供了很强大的日志功能，性能高，内存分配次数少，适合对日志性能要求 很高的项目。另外，zap 包中的子包 zapcore，提供了很多底层的日志接口，适合用来 做二次封装。

举个自己选择日志包来进行二次开发的例子：在做容器云平台开发时，发现 Kubernetes 源码中大量使用了 glog，这时就需要日志包能够兼容 glog。于是，基于 zap 和 zapcore 封装了github.com/marmotedu/iam/pkg/log日志包，这个日志包可 以很好地兼容 glog。 

在实际项目开发中，可以根据项目需要，从上面几个日志包中进行选择，直接使用，但 更多时候，还需要基于这些包来进行定制开发。为了更深入地掌握日志包的设计和 开发，接下来，会从 0 到 1 开发一个日志包。 

### 从 0 编写一个日志包 

接下来，会展示如何快速编写一个具备基本功能的日志包，通过这个简短的日志包实现掌握日志包的核心设计思路。该日志包主要实现以下几个功能：

- 支持自定义配置。 
- 支持文件名和行号。 
- 支持日志级别 Debug、Info、Warn、Error、Panic、Fatal。 
- 支持输出到本地文件和标准输出。
- 支持 JSON 和 TEXT 格式的日志输出，支持自定义日志格式。 
- 支持选项模式。

日志包名称为cuslog，示例项目完整代码存放在 cuslog。 具体实现分为以下四个步骤：

- 定义：定义日志级别和日志选项。 
- 创建：创建 Logger 及各级别日志打印方法。 
- 写入：将日志输出到支持的输出中。 
- 自定义：自定义日志输出格式。

#### 定义日志级别和日志选项 

一个基本的日志包，首先需要定义好日志级别和日志选项。本示例将定义代码保存在 options.go文件中。 可以通过如下方式定义日志级别：

```go
package cuslog

type Level uint8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

var LevelNameMapping = map[Level]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	PanicLevel: "PANIC",
	FatalLevel: "FATAL",
}
```

在日志输出时，要通过对比开关级别和输出级别的大小，来决定是否输出，所以日志级别 Level 要定义成方便比较的数值类型。几乎所有的日志包都是用常量计数器 iota 来定义日志级别。 

另外，因为要在日志输出中，输出可读的日志级别（例如输出 INFO 而不是 1），所以需 要有 Level 到 Level Name 的映射 LevelNameMapping，LevelNameMapping 会在格 式化时用到。 

接下来看定义日志选项。日志需要是可配置的，方便开发者根据不同的环境设置不同的日 志行为，比较常见的配置选项为：

- 日志级别。 
- 输出位置，例如标准输出或者文件。 
- 输出格式，例如 JSON 或者 Text。 是
- 否开启文件名和行号。

本示例的日志选项定义如下：

```go
type options struct {
	output        io.Writer
	level         Level
	stdLevel      Level
	formatter     Formatter
	disableCaller bool
}
```

为了灵活地设置日志的选项，可以通过选项模式，来对日志选项进行设置：

```go
type Option func(*options)

func initOptions(opts ...Option) (o *options) {
   o = &options{}
   for _, opt := range opts {
      opt(o)
   }
   if o.output == nil {
      o.output = os.Stderr
   }
   if o.formatter == nil {
      o.formatter = &TextFormatter{}
   }
   return
}
func WithLevel(level Level) Option {
   return func(o *options) {
      o.level = level
   }
}

// ...

func SetOptions(opts ...Option) {
   std.SetOptions(opts...)
}
func (l *logger) SetOptions(opts ...Option) {
   l.mu.Lock()
   defer l.mu.Unlock()
   for _, opt := range opts {
      opt(l.opt)
   }
}
```

具有选项模式的日志包，可通过以下方式，来动态地修改日志的选项：

```go
cuslog.SetOptions(cuslog.WithLevel(cuslog.DebugLevel))
```

可以根据需要，对每一个日志选项创建设置函数 WithXXXX 。这个示例日志包支持如下 选项设置函数：

- WithOutput（output io.Writer）：设置输出位置。 
- WithLevel（level Level）：设置输出级别。 
- WithFormatter（formatter Formatter）：设置输出格式。 
- WithDisableCaller（caller bool）：设置是否打印文件名和行号。

#### 创建 Logger 及各级别日志打印方法 

为了打印日志，需要根据日志配置，创建一个 Logger，然后通过调用 Logger 的日志 打印方法，完成各级别日志的输出。本示例将创建代码保存在 logger.go文件中。 

可以通过如下方式创建 Logger：

```go
var std = New()

type logger struct {
   opt       *options
   mu        sync.Mutex
   entryPool *sync.Pool
}

func New(opts ...Option) *logger {
   logger := &logger{opt: initOptions(opts...)}
   logger.entryPool = &sync.Pool{New: func() interface{} {
      return entry(logger)
   }}
   return logger
}
```

上述代码中，定义了一个 Logger，并实现了创建 Logger 的 New 函数。日志包都会有一 个默认的全局 Logger，本示例通过 var std = New() 创建了一个全局的默认 Logger。

cuslog.Debug、cuslog.Info 和 cuslog.Warnf 等函数，则是通过调用 std Logger 所提供的方法来打印日志的。 

定义了一个 Logger 之后，还需要给该 Logger 添加最核心的日志打印方法，要提供所有支持级别的日志打印方法。 

如果日志级别是 Xyz，则通常需要提供两类方法，分别是非格式化方法Xyz(args ...interface{})和格式化方法Xyzf(format string, args ...interface{})，例如：

```go
func (l *logger) Debug(args ...interface{}) {
   l.entry().write(DebugLevel, FmtEmptySeparate, args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
   l.entry().write(DebugLevel, format, args...)
}
```

本示例实现了如下方法：Debug、Debugf、Info、Infof、Warn、Warnf、Error、 Errorf、Panic、Panicf、Fatal、Fatalf。更详细的实现，可以参考 cuslog/logger.go。 

这里要注意，Panic、Panicf 要调用 panic() 函数，Fatal、Fatalf 函数要调用 os.Exit(1) 函数。 

#### 将日志输出到支持的输出中 

调用日志打印函数之后，还需要将这些日志输出到支持的输出中，所以需要实现 write 函 数，它的写入逻辑保存在 entry.go文件中。实现方式如下：

```go
type Entry struct {
   logger *logger
   Buffer *bytes.Buffer
   Map    map[string]interface{}
   Level  Level
   Time   time.Time
   File   string
   Line   int
   Func   string
   Format string
   Args   []interface{}
}

func (e *Entry) write(level Level, format string, args ...interface{}) {
   if e.logger.opt.level > level {
      return
   }
   e.Time = time.Now()
   e.Level = level
   e.Format = format
   e.Args = args
   if !e.logger.opt.disableCaller {
      if pc, file, line, ok := runtime.Caller(2); !ok {
         e.File = "???"
         e.Func = "???"
      } else {
         e.File, e.Line, e.Func = file, line, runtime.FuncForPC(pc).Name()
         e.Func = e.Func[strings.LastIndex(e.Func, "/")+1:]
      }
   }
   e.format()
   e.writer()
   e.release()
}

func (e *Entry) format() {
   _ = e.logger.opt.formatter.Format(e)
}

func (e *Entry) writer() {
   e.logger.mu.Lock()
   _, _ = e.logger.opt.output.Write(e.Buffer.Bytes())
   e.logger.mu.Unlock()
}

func (e *Entry) release() {
   e.Args, e.Line, e.File, e.Format, e.Func = nil, 0, "", "", ""
   e.Buffer.Reset()
   e.logger.entryPool.Put(e)
}
```

上述代码，首先定义了一个 Entry 结构体类型，该类型用来保存所有的日志信息，即日志 配置和日志内容。写入逻辑都是围绕 Entry 类型的实例来完成的。 

用 Entry 的 write 方法来完成日志的写入，在 write 方法中，会首先判断日志的输出级别 和开关级别，如果输出级别小于开关级别，则直接返回，不做任何记录。 

在 write 中，还会判断是否需要记录文件名和行号，如果需要则调用 runtime.Caller() 来获取文件名和行号，调用 runtime.Caller() 时，要注意传入 正确的栈深度。 

write 函数中调用 e.format 来格式化日志，调用 e.writer 来写入日志，在创建 Logger 传入的日志配置中，指定了输出位置 output io.Writer ，output 类型为 io.Writer ，示例如下：

```go
type Writer interface {
	Write(p []byte) (n int, err error)
}
```

io.Writer 实现了 Write 方法可供写入，所以只需要调用 `e.logger.opt.output.Write(e.Buffer.Bytes()) `即可将日志写入到指定的位置 中。

最后，会调用 release() 方法来清空缓存和对象池。至此，就完成了日志的记录和 写入。

#### 自定义日志输出格式 

cuslog 包支持自定义输出格式，并且内置了 JSON 和 Text 格式的 Formatter。 Formatter 接口定义为：

```go
type Formatter interface {
   Format(entry *Entry) error
}
```

cuslog 内置的 Formatter 有两个：JSON和TEXT。

#### 测试日志包 

cuslog 日志包开发完成之后，可以编写测试代码，调用 cuslog 包来测试 cuslog 包，代 码如下：

```go
import (
	"log"
	"os"

	"github.com/marmotedu/gopractise-demo/log/cuslog"
)

func main() {
	cuslog.Info("std log")
	cuslog.SetOptions(cuslog.WithLevel(cuslog.DebugLevel))
	cuslog.Debug("change std log to debug level")
	cuslog.SetOptions(cuslog.WithFormatter(&cuslog.JsonFormatter{IgnoreBasicFields: false}))
	cuslog.Debug("log in json format")
	cuslog.Info("another log in json format")

	// 输出到文件
	fd, err := os.OpenFile("test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("create file test.log failed")
	}
	defer fd.Close()

	l := cuslog.New(cuslog.WithLevel(cuslog.InfoLevel),
		cuslog.WithOutput(fd),
		cuslog.WithFormatter(&cuslog.JsonFormatter{IgnoreBasicFields: false}),
	)
	l.Info("custom log with json formatter")
}
```

将上述代码保存在 main.go 文件中，运行：

```sh
$ go run main/main.go
2021-11-24T02:20:37+08:00 INFO main.go:10 std log
2021-11-24T02:20:37+08:00 DEBUG main.go:12 change std log to debug level
{"level":"DEBUG","time":"2021-11-24T02:20:37+08:00","file":"/Users/rmliu/CodeStudy/MyGo/Code/src/cuslog/example/main/main.go:15","func":"main.main","message":"log in json format"}
{"time":"2021-11-24T02:20:37+08:00","file":"/Users/rmliu/CodeStudy/MyGo/Code/src/cuslog/example/main/main.go:16","func":"main.main","message":"another log in json format","level":"INFO"}
```

到这里日志包就开发完成了，完整包见 log/cuslog。

### IAM 项目日志包设计 

这一讲的最后，再来看下 IAM 项目中，日志包是怎么设计的。 

#### log 包的存放位置

先来看一下 IAM 项目 log 包的存放位置：pkg/log。放在这个位置，主要有两个原因： 

- 第一个，log 包属于 IAM 项目，有定制开发的内容；
- 第二个，log 包功能完备、成熟，外部项目也可以使用。 

#### log 包的 Options

该 log 包是基于 go.uber.org/zap 包封装而来的，根据需要添加了更丰富的功能。接下 来，通过 log 包的 Options，来看下 log 包所实现的功能：

```go
// Options contains configuration items related to log.
type Options struct {
	OutputPaths       []string `json:"output-paths"       mapstructure:"output-paths"`
	ErrorOutputPaths  []string `json:"error-output-paths" mapstructure:"error-output-paths"`
	Level             string   `json:"level"              mapstructure:"level"`
	Format            string   `json:"format"             mapstructure:"format"`
	DisableCaller     bool     `json:"disable-caller"     mapstructure:"disable-caller"`
	DisableStacktrace bool     `json:"disable-stacktrace" mapstructure:"disable-stacktrace"`
	EnableColor       bool     `json:"enable-color"       mapstructure:"enable-color"`
	Development       bool     `json:"development"        mapstructure:"development"`
	Name              string   `json:"name"               mapstructure:"name"`
}
```

Options 各配置项含义如下：

- development：是否是开发模式。如果是开发模式，会对 DPanicLevel 进行堆栈跟 踪。 
- name：Logger 的名字。 
- disable-caller：是否开启 caller，如果开启会在日志中显示调用日志所在的文件、函数 和行号。 
- disable-stacktrace：是否在 Panic 及以上级别禁止打印堆栈信息。 
- enable-color：是否开启颜色输出，true，是；false，否。 
- level：日志级别，优先级从低到高依次为：Debug, Info, Warn, Error, Dpanic, Panic, Fatal。 
- format：支持的日志输出格式，目前支持 Console 和 JSON 两种。Console 其实就是 Text 格式。 
- output-paths：支持输出到多个输出，用逗号分开。支持输出到标准输出（stdout）和 文件。 
- error-output-paths：zap 内部 (非业务) 错误日志输出路径，多个输出，用逗号分开。

log 包的 Options 结构体支持以下 3 个方法：

- Build 方法。Build 方法可以根据 Options 构建一个全局的 Logger。 
- AddFlags 方法。AddFlags 方法可以将 Options 的各个字段追加到传入的 pflag.FlagSet 变量中。 
- String 方法。String 方法可以将 Options 的值以 JSON 格式字符串返回。

#### 3 种日志记录方法

log 包实现了以下 3 种日志记录方法：

```go
log.Info("This is a info message", log.Int32("int_key", 10))
log.Infof("This is a formatted %s message", "info")
log.Infow("Message printed with Infow", "X-Request-ID", "fbf54504-64da-4088-9b86-67824a7fb508")
```

- Info 使用指定的 key/value 记录日志。
- Infof 格式化记录日志。 
- Infow 也是使用指定的 key/value 记录日志，跟 Info 的区别是：使用 Info 需要指定值的类型，通过指定值的 日志类型，日志库底层不需要进行反射操作，所以使用 Info 记录日志性能最高。 

log 包支持非常丰富的类型，具体可以参考 types.go。 

上述日志输出为：

```sh
2021-07-06 14:02:07.070 INFO This is a info message {"int_key": 10}
2021-07-06 14:02:07.071 INFO This is a formatted info message
2021-07-06 14:02:07.071 INFO Message printed with Infow {"X-Request-ID": "fbf54504-64da-4088-9b86-67824a7fb508")
```

log 包为每种级别的日志都提供了 3 种日志记录方式，举个例子：假设日志格式为 Xyz ， 则分别提供了 `Xyz(msg string, fields ...Field)` ，`Xyzf(format string, v ...interface{}) `，`Xyzw(msg string, keysAndValues ...interface{})` 3 种日志记录方法。 

另外，log 包相较于一般的日志包，还提供了众多记录日志的方法。 

#### log 包支持 V Level

第一个方法， log 包支持 V Level，可以通过整型数值来灵活指定日志级别，数值越大， 优先级越低。例如：

```go
func main() {
	defer log.Flush()

	log.V(0).Info("This is a V level message")
	log.V(0).Infow("This is a V level message with fields", "X-Request-ID", "7a7b9f24-4cae-4b2a-9464-69088b45b904")

	// V level使用
	log.V(1).Info("This is a V level message")
	log.V(1).Infof("This is a %s V level message", "formatted")
	log.V(1).Infow("This is a V level message with fields", "X-Request-ID", "7a7b9f24-4cae-4b2a-9464-69088b45b904")
}
```

这里要注意，Log.V 只支持 Info 、Infof 、Infow三种日志记录方法。

####  log 包支持 WithValues 函数

第二个方法， log 包支持 WithValues 函数，例如：

```go
// WithValues使用
lv := log.WithValues("X-Request-ID", "7a7b9f24-4cae-4b2a-9464-69088b45b904")
lv.Infow("Info message printed with [WithValues] logger")
lv.Infow("Debug message printed with [WithValues] logger")
```

上述日志输出如下：

```sh
2021-11-24 02:50:41.609 INFO    Info message printed with [WithValues] logger   {"X-Request-ID": "7a7b9f24-4cae-4b2a-9464-69088b45b904"}
2021-11-24 02:50:41.609 INFO    Debug message printed with [WithValues] logger  {"X-Request-ID": "7a7b9f24-4cae-4b2a-9464-69088b45b904"}
```

WithValues 可以返回一个携带指定 key-value 的 Logger，供后面使用。 

#### log 包提供 WithContext 和 FromContext

第三个方法， log 包提供 WithContext 和 FromContext 用来将指定的 Logger 添加到 某个 Context 中，以及从某个 Context 中获取 Logger，例如：

```go
// Context使用
ctx := lv.WithContext(context.Background())
lc := log.FromContext(ctx)
lc.Info("Message printed with [WithContext] logger")
```

WithContext和FromContext非常适合用在以context.Context传递的函数中，例 如：

```go
// example/context/main.go

func main() {
	
  // ...

	// WithValues使用
	lv := log.WithValues("X-Request-ID", "7a7b9f24-4cae-4b2a-9464-69088b45b904")

	// Context使用
	lv.Infof("Start to call pirntString function")
	ctx := lv.WithContext(context.Background())
	pirntString(ctx, "World")
}

func pirntString(ctx context.Context, str string) {
	lc := log.FromContext(ctx)
	lc.Infof("Hello %s", str)
}
```

上述代码输出如下：

```sh
2021-11-24 02:45:57.293 INFO    Start to call pirntString function      {"X-Request-ID": "7a7b9f24-4cae-4b2a-9464-69088b45b904"}
2021-11-24 02:45:57.294 INFO    Hello World     {"X-Request-ID": "7a7b9f24-4cae-4b2a-9464-69088b45b904"}
```

将 Logger 添加到 Context 中，并通过 Context 在不同函数间传递，可以使 key-value 在不同函数间传递。

例如上述代码中， X-Request-ID 在 main 函数、printString 函数 中的日志输出中均有记录，从而实现了一种调用链的效果。 

#### 从 Context 中提取出指定的 key-value

第四个方法， 可以很方便地从 Context 中提取出指定的 key-value，作为上下文添加到日 志输出中，例如 internal/apiserver/api/v1/user/create.go文件中的日志调用：

```sh
log.L(c).Info("user create function called.")
```

通过调用 Log.L() 函数，实现如下：

```go
// L method output with specified context value.
func L(ctx context.Context) *zapLogger {
	return std.L(ctx)
}
func (l *zapLogger) L(ctx context.Context) *zapLogger {
	lg := l.clone()
	requestID, _ := ctx.Value(KeyRequestID).(string)
	username, _ := ctx.Value(KeyUsername).(string)
	lg.zapLogger = lg.zapLogger.With(zap.String(KeyRequestID, requestID), zap.String(KeyUsername, username))
	return lg
}
```

L() 方法会从传入的 Context 中提取出 requestID 和 username ，追加到 Logger 中， 并返回 Logger。这时候调用该 Logger 的 Info、Infof、Infow 等方法记录日志，输出的 日志中均包含 requestID 和 username 字段，例如：

```sh
2021-07-06 14:46:00.743 INFO apiserver secret/create.go:23      create secret function called.  {"requestID": "73144bed-534d-4f68-8e8d-dc8a8ed48507", "username": "admin"}
```

通过将 Context 在函数间传递，很容易就能实现调用链效果，例如：

```go
// Create add new secret key pairs to the storage.
func (s *SecretHandler) Create(c *gin.Context) {
	log.L(c).Info("create secret function called.")
  
	...
  
	sec, err := s.store.Secrets().List(c, username, metav1.ListOptions{
		Offset: pointer.ToInt64(0),
		Limit: pointer.ToInt64(-1),
	})
  
	...
  
	if err := s.srv.Secrets().Create(c, &r, metav1.CreateOptions{}); err != ni
	core.WriteResponse(c, err, nil)
	return
}
```

上述代码输出为：

```sh
2021-07-06 14:46:00.743 INFO apiserver secret/create.go:23 secret/create.go:23      create secret function called.  {"requestID": "73144bed-534d-4f68-8e8d-dc8a8ed48507", "username": "admin"}
2021-07-06 14:46:00.744 INFO apiserver secret/create.go:23 list  secret ...
2021-07-06 14:46:00.745 INFO apiserver secret/create.go:23 insert secret ...
```

这里要注意， log.L 函数默认会从 Context 中取 requestID 和 username 键，这跟 IAM 项目有耦合度，但这不影响 log 包供第三方项目使用。这也是建议自己封装日志 包的原因。 

### 总结 

开发一个日志包，很多时候需要基于一些业界优秀的开源日志包进行二次开发。当前 很多项目的日志包都是基于 zap 日志包来封装的，如果有封装的需要，建议优先选 择 zap 日志包。 

这一讲中，先介绍了标准库 log 包、glog、logrus 和 zap 这四种常用的日志包， 然后展现了开发一个日志包的四个步骤，步骤如下：

- 定义日志级别和日志选项。 
- 创建 Logger 及各级别日志打印方法。 
- 将日志输出到支持的输出中。
- 自定义日志输出格式。

最后，介绍了 IAM 项目封装的 log 包的设计和使用方式。log 包基于 go.uber.org/zap封装，并提供了以下强大特性：

- log 包支持 V Level，可以灵活的通过整型数值来指定日志级别。 
- log 包支持 WithValues 函数， WithValues 可以返回一个携带指定 key-value 对的 Logger，供后面使用。 
- log 包提供 WithContext 和 FromContext 用来将指定的 Logger 添加到某个 Context 中和从某个 Context 中获取 Logger。 
- log 包提供了 Log.L() 函数，可以很方便的从 Context 中提取出指定的 key-value 对，作为上下文添加到日志输出中。

### 课后练习

- 尝试实现一个新的 Formatter，可以使不同日志级别以不同颜色输出（例如：Error 级 别的日志输出中 Error 字符串用红色字体输出， Info 字符串用白色字体输出）。
- 尝试将 runtime.Caller(2)函数调用中的 2 改成 1 ，看看日志输出是否跟修改前有差 异，如果有差异，思考差异产生的原因。



## GO 项目之构建三剑客：Pflag、Viper、Cobra

来聊聊构建应用时常用的 Go 包。 

因为 IAM 项目使用了 Pflag、Viper 和 Cobra 包来构建 IAM 的应用框架，这里简单介绍下这 3 个包的核心功能和使用方式。其实如果单独讲每个包 的话，还是有很多功能可讲的，但这一讲的目的是减小后面学习 IAM 源码的难度， 所以会主要介绍跟 IAM 相关的功能。 

在正式介绍这三个包之前，先来看下如何构建应用的框架。 

### 如何构建应用框架

想知道如何构建应用框架，首先要明白，一个应用框架包含哪些部分。一个 应用框架需要包含以下 3 个部分：

- 命令行参数解析：主要用来解析命令行参数，这些命令行参数可以影响命令的运行效 果。 
- 配置文件解析：一个大型应用，通常具有很多参数，为了便于管理和配置这些参数，通 常会将这些参数放在一个配置文件中，供程序读取并解析。 
- 应用的命令行框架：应用最终是通过命令来启动的。这里有 3 个需求点，
  - 一是命令需要 具备 Help 功能，这样才能告诉使用者如何去使用；
  - 二是命令需要能够解析命令行参数 和配置文件；
  - 三是命令需要能够初始化业务代码，并最终启动业务进程。
  - 也就是说，命令需要具备框架的能力，来纳管这 3 个部分。

这 3 个部分的功能，可以自己开发，也可以借助业界已有的成熟实现。跟之前的想法一 样，不建议自己开发，更建议采用业界已有的成熟实现。

命令行参数可以通过 Flag来解析，配置文件可以通过 Viper来解析，应用的命令行框架则可以通过 Cobra 来实现。

这 3 个包目前也是最受欢迎的包，并且这 3 个包不是割裂的，而是有联系的，可以有机地组合这 3 个包，从而实现一个非常强大、优秀的应用命令行框架。 

接下来，来详细看下，这 3 个包在 Go 项目开发中是如何使用的。 

### 命令行参数解析工具：Pflag 使用介绍 

Go 服务开发中，经常需要给开发的组件加上各种启动参数来配置服务进程，影响服务的行 为。

像 kube-apiserver 就有多达 200 多个启动参数，而且这些参数的类型各不相同（例 如：string、int、ip 类型等），使用方式也不相同（例如：需要支持--长选项，-短选项 等），所以需要一个强大的命令行参数解析工具。 

虽然 Go 源码中提供了一个标准库 Flag 包，用来对命令行参数进行解析，但在大型项目中 应用更广泛的是另外一个包：Pflag。

Pflag 提供了很多强大的特性，非常适合用来构建大 型项目，一些耳熟能详的开源项目都是用 Pflag 来进行命令行参数解析的，例如： Kubernetes、Istio、Helm、Docker、Etcd 等。 

接下来，就来介绍下如何使用 Pflag。Pflag 主要是通过创建 Flag 和 FlagSet 来使用 的。

#### Pflag 包 Flag 定义 

Pflag 可以对命令行参数进行处理，一个命令行参数在 Pflag 包中会解析为一个 Flag 类型 的变量。

Flag 是一个结构体，定义如下：

```go
// A Flag represents the state of a flag.
type Flag struct {
   Name                string              // name as it appears on command line  // flag长选项的名称
   Shorthand           string              // one-letter abbreviated flag   // flag短选项的名称，一个缩写的字符
   Usage               string              // help message  // flag的使用文本
   Value               Value               // value as set  // flag的值
   DefValue            string              // default value (as text); for usage message  // flag的默认值
   Changed             bool                // If the user set the value (or if left to default)  // 记录flag的值是否有被设置过
   NoOptDefVal         string              // default value (as text); if the flag is on the command line without any options  // 当flag出现在命令行，但是没有指定选项值时的默认值
   Deprecated          string              // If this flag is deprecated, this string is the new or now thing to use  // 记录该flag是否被放弃
   Hidden              bool                // used by cobra.Command to allow flags to be hidden from help/usage text  // 如果值为true，则从help/usage输出信息中隐藏该flag
   ShorthandDeprecated string              // If the shorthand of this flag is deprecated, this string is the new or now thing to use  // 如果flag的短选项被废弃，当使用flag的短选项时打印该信息
   Annotations         map[string][]string // used by cobra.Command bash autocomple code  // 给flag设置注解
}
```

Flag 的值是一个 Value 类型的接口，Value 定义如下：

```go
// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
type Value interface {
   String() string  // 将flag类型的值转换为string类型的值，并返回string的内容
   Set(string) error   // 将string类型的值转换为flag类型的值，转换失败报错
   Type() string  // 返回flag的类型，例如：string、int、ip等
}
```

通过将 Flag 的值抽象成一个 interface 接口，就可以自定义 Flag 的类型了。只要实 现了 Value 接口的结构体，就是一个新类型。 

#### Pflag 包 FlagSet 定义 

Pflag 除了支持单个的 Flag 之外，还支持 FlagSet。

FlagSet 是一些预先定义好的 Flag 的 集合，几乎所有的 Pflag 操作，都需要借助 FlagSet 提供的方法来完成。在实际开发中， 可以使用两种方法来获取并使用 FlagSet：

- 方法一，调用 NewFlagSet 创建一个 FlagSet。
- 方法二，使用 Pflag 包定义的全局 FlagSet：CommandLine。实际上 CommandLine 也是由 NewFlagSet 函数创建的。

先来看下第一种方法，自定义 FlagSet。下面是一个自定义 FlagSet 的示例：

```go
var version bool
flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
flagSet.BoolVar(&version, "version", true, "Print version information and quit.")
```

可以通过定义一个新的 FlagSet 来定义命令及其子命令的 Flag。 

再来看下第二种方法，使用全局 FlagSet。下面是一个使用全局 FlagSet 的示例：

```go
import (
	"github.com/spf13/pflag"
)

pflag.BoolVarP(&version, "version", "v", true, "Print version information and quit.")
```

这其中，pflag.BoolVarP 函数定义如下：

```go
// BoolVarP is like BoolVar, but accepts a shorthand letter that can be used after a single dash.
func BoolVarP(p *bool, name, shorthand string, value bool, usage string) {
	flag := CommandLine.VarPF(newBoolValue(value, p), name, shorthand, usage)
	flag.NoOptDefVal = "true"
}
```

可以看到 pflag.BoolVarP 最终调用了 CommandLine，CommandLine 是一个包级别的 变量，定义为：

```go
// CommandLine is the default set of command-line flags, parsed from os.Args.
var CommandLine = NewFlagSet(os.Args[0], ExitOnError)
```

在一些不需要定义子命令的命令行工具中，可以直接使用全局的 FlagSet，更加简单方 便。 

#### Pflag 使用方法 

上面，介绍了使用 Pflag 包的两个核心结构体。

接下来，详细介绍下 Pflag 的常 见使用方法。Pflag 有很多强大的功能，这里介绍 7 个常见的使用方法。

##### 多种命令行参数定义方式

1. 支持多种命令行参数定义方式。

Pflag 支持以下 4 种命令行参数定义方式：

- 支持长选项、默认值和使用文本，并将标志的值存储在指针中。

  - ```go
    var name = pflag.String("name", "colin", "Input Your Name")
    ```

- 支持长选项、短选项、默认值和使用文本，并将标志的值存储在指针中。

  - ```go
     var name = pflag.StringP("name", "n", "colin", "Input Your Name")
    ```

- 支持长选项、默认值和使用文本，并将标志的值绑定到变量。

  - ```go
    var name string
    pflag.StringVar(&name, "name", "colin", "Input Your Name")
    ```

- 支持长选项、短选项、默认值和使用文本，并将标志的值绑定到变量。

  - ```go
    var name string
    pflag.StringVarP(&name, "name", "n","colin", "Input Your Name")
    ```

上面的函数命名是有规则的：

- 函数名带Var说明是将标志的值绑定到变量，否则是将标志的值存储在指针中。 
- 函数名带P说明支持短选项，否则不支持短选项。

##### Get 获取参数的值

2. 使用Get获取参数的值。

可以使用Get来获取标志的值，代表 Pflag 所支持的类型。

例如：有一个 pflag.FlagSet，带有一个名为 flagname 的 int 类型的标志，可以使用GetInt()来获取 int 值。需要注意 flagname 必须存在且必须是 int，例如：

```go
i, err := flagset.GetInt("flagname")
```

##### 获取非选项参数

3. 获取非选项参数。

代码示例如下：

```go
import (
	"fmt"

	"github.com/spf13/pflag"
)

var (
	flagvar = pflag.Int("flagname", 1234, "help message for flagname")
)

func main() {
	pflag.Parse()

	fmt.Printf("argument number is: %v\n", pflag.NArg())
	fmt.Printf("argument list is: %v\n", pflag.Args())
	fmt.Printf("the first argument is: %v\n", pflag.Arg(0))
}
```

执行上述代码，输出如下：

```sh
$ go run example2.go arg1 arg2 
argument number is: 2
argument list is: [arg1 arg2]
the first argument is: arg1
```

在定义完标志之后，可以调用 pflag.Parse() 来解析定义的标志。解析后，可通过 pflag.Args() 返回所有的非选项参数，通过pflag.Arg(i)返回第 i 个非选项参数。参 数下标 0 到 pflag.NArg() - 1。

##### 指定了选项但是没有指定选项值时的默认值

4. 指定了选项但是没有指定选项值时的默认值。

创建一个 Flag 后，可以为这个 Flag 设置 pflag.NoOptDefVal。如果一个 Flag 具有 NoOptDefVal，并且该 Flag 在命令行上没有设置这个 Flag 的值，则该标志将设置为 NoOptDefVal 指定的值。例如：

```go
var ip = flag.IntP("flagname", "f", 1234, "help message")
flag.Lookup("flagname").NoOptDefVal = "4321"
```

上面的代码会产生结果，具体可以参照下表：

![image-20211124211639949](IAM-document.assets/image-20211124211639949.png)

##### 弃用标志或标志的简写

5. 弃用标志或者标志的简写。

Pflag 可以弃用标志或者标志的简写。

弃用的标志或标志简写在帮助文本中会被隐藏，并在 使用不推荐的标志或简写时打印正确的用法提示。例如，弃用名为 logmode 的标志，并告知用户应该使用哪个标志代替：

```go
// deprecate a flag by specifying its name and a usage message
pflag.CommandLine.MarkDeprecated("logmode", "please use --log-mode instead")
```

这样隐藏了帮助文本中的 logmode，并且当使用 logmode 时，打印了Flag -- logmode has been deprecated, please use --log-mode instead。 

##### 保留名为 port 的标志但是弃用简写形式

6) 保留名为 port 的标志，但是弃用它的简写形式。

```go
pflag.IntVarP(&port, "port", "P", 3306, "MySQL service host port.")
// deprecate a flag shorthand by specifying its flag name and a usage message
pflag.CommandLine.MarkShorthandDeprecated("port", "please use --port only")
```

这样隐藏了帮助文本中的简写 P，并且当使用简写 P 时，打印了Flag shorthand -P has been deprecated, please use --port only。usage message 在此处必不 可少，并且不应为空。

##### 隐藏标志

7. 隐藏标志。

可以将 Flag 标记为隐藏的，这意味着它仍将正常运行，但不会显示在 usage/help 文本 中。

例如：隐藏名为 secretFlag 的标志，只在内部使用，并且不希望它显示在帮助文本或 者使用文本中。代码如下：

```go
// hide a flag by specifying its name
pflag.CommandLine.MarkHidden("secretFlag")
```

至此，介绍了 Pflag 包的重要用法。接下来，再来看下如何解析配置文件。 

### 配置解析神器：Viper 使用介绍

几乎所有的后端服务，都需要一些配置项来配置服务，一些小型的项目，配置不是 很多，可以选择只通过命令行参数来传递配置。

但是大型项目配置很多，通过命令行参数 传递就变得很麻烦，不好维护。标准的解决方案是将这些配置信息保存在配置文件中，由 程序启动时加载和解析。

Go 生态中有很多包可以加载并解析配置文件，目前最受欢迎的是 Viper 包。 

Viper 是 Go 应用程序现代化的、完整的解决方案，能够处理不同格式的配置文件，在构建现代应用程序时，不必担心配置文件格式。

Viper 也能够满足对应用配置的各 种需求。 

Viper 可以从不同的位置读取配置，不同位置的配置具有不同的优先级，高优先级的配置 会覆盖低优先级相同的配置，按优先级从高到低排列如下：

-  通过 viper.Set 函数显示设置的配置
- 命令行参数 
- 环境变量 
- 配置文件 
- Key/Value 存储 
- 默认值

这里需要注意，Viper 配置键不区分大小写。 

Viper 有很多功能，最重要的两类功能是读入配置和读取配置，Viper 提供不同的方式来实 现这两类功能。接下来，就来详细介绍下 Viper 如何读入配置和读取配置。 

#### 读入配置 

读入配置，就是将配置读入到 Viper 中，有如下读入方式：

- 设置默认的配置文件名。 
- 读取配置文件。 
- 监听和重新读取配置文件。
- 从 io.Reader 读取配置。 
- 从环境变量读取。 
- 从命令行标志读取。 
- 从远程 Key/Value 存储读取。

这几个方法的具体读入方式，可以看下面的展示。

##### 设置默认值

1. 设置默认值。

一个好的配置系统应该支持默认值。Viper 支持对 key 设置默认值，当没有通过配置文 件、环境变量、远程配置或命令行标志设置 key 时，设置默认值通常是很有用的，可以让 程序在没有明确指定配置时也能够正常运行。例如：

```go
viper.SetDefault("ContentDir", "content")
viper.SetDefault("LayoutDir", "layouts")
viper.SetDefault("Taxonomies", map[string]string{"tag": "tags", "category": "categories"})
```

##### 读取配置文件

2. 读取配置文件。

Viper 可以读取配置文件来解析配置，支持 JSON、TOML、YAML、YML、Properties、 Props、Prop、HCL、Dotenv、Env 格式的配置文件。

Viper 支持搜索多个路径，并且默 认不配置任何搜索路径，将默认决策留给应用程序。 以下是如何使用 Viper 搜索和读取配置文件的示例：

```go
package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfg   = pflag.StringP("config", "c", "", "Configuration file.")
	help  = pflag.BoolP("help", "h", false, "Show this help message.")
)

func main() {
	pflag.Parse()
	if *help {
		pflag.Usage()
		return
	}

	// 从配置文件中读取配置
	if *cfg != "" {
		viper.SetConfigFile(*cfg)   // 指定配置文件名
		viper.SetConfigType("yaml") // 如果配置文件名中没有文件扩展名，则需要指定配置文件的格式，告诉viper以何种格式解析文件
	} else {
		viper.AddConfigPath(".")          // 把当前目录加入到配置文件的搜索路径中
		viper.AddConfigPath("$HOME/.iam") // 配置文件搜索路径，可以设置多个配置文件搜索路径
		viper.SetConfigName("config")     // 配置文件名称（没有文件扩展名）
	}

	if err := viper.ReadInConfig(); err != nil { // 读取配置文件。如果指定了配置文件名，则使用指定的配置文件，否则在注册的搜索路径中搜索
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// 打印当前使用的配置文件名
	fmt.Printf("Used configuration file is: %s\n", viper.ConfigFileUsed())
}
```

Viper 支持设置多个配置文件搜索路径(AddConfigPath)，需要注意添加搜索路径的顺序，Viper 会根据添加 的路径顺序搜索配置文件，如果找到则停止搜索。

如果调用 SetConfigFile 直接指定了配 置文件名，并且配置文件名没有文件扩展名时，需要显式指定配置文件的格式，以使 Viper 能够正确解析配置文件。 

如果通过搜索的方式查找配置文件，则需要注意，SetConfigName 设置的配置文件名是 不带扩展名的，在搜索时 Viper 会在文件名之后追加文件扩展名，并尝试搜索所有支持的 扩展类型。

##### 监听和重新读取配置文件

3. 监听和重新读取配置文件。

Viper 支持在运行时让应用程序实时读取配置文件，也就是热加载配置。

可以通过 WatchConfig 函数热加载配置。在调用 WatchConfig 函数之前，需要确保已经添加了配 置文件的搜索路径。

另外，还可以为 Viper 提供一个回调函数，以便在每次发生更改时运 行。这里也给个示例：

```go
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
  // 配置文件发生变更之后会调用的回调函数
	fmt.Println("Config file changed:", e.Name)
})
```

不建议在实际开发中使用热加载功能，因为即使配置热加载了，程序中的代码也不一定 会热加载。

例如：修改了服务监听端口，但是服务没有重启，这时候服务还是监听在老的 端口上，会造成不一致。 

##### 设置配置值

4) 设置配置值。 

可以通过 viper.Set() 函数来显式设置配置：

```go
viper.Set("user.username", "colin")
```

##### 使用环境变量

5. 使用环境变量。

Viper 还支持环境变量，通过如下 5 个函数来支持环境变量：

- AutomaticEnv() 
- BindEnv(input …string) error 
- SetEnvPrefix(in string) 
- SetEnvKeyReplacer(r *strings.Replacer) 
- AllowEmptyEnv(allowEmptyEnv bool)

这里要注意：Viper 读取环境变量是区分大小写的。

Viper 提供了一种机制来确保 Env 变 量是唯一的。

通过使用 SetEnvPrefix，可以告诉 Viper 在读取环境变量时使用前缀。 BindEnv 和 AutomaticEnv 都将使用此前缀。比如，设置了 viper.SetEnvPrefix(“VIPER”)，当使用 viper.Get(“apiversion”) 时，实际读取的环境 变量是VIPER_APIVERSION。

BindEnv 需要一个或两个参数。第一个参数是键名，第二个是环境变量的名称，环境变量的名称区分大小写。

- 如果未提供 Env 变量名，则 Viper 将假定 Env 变量名为：环境变量前缀 _ 键名全大写。例如：前缀为 VIPER，key 为 username，则 Env 变量名为 VIPER_USERNAME。
- 当显示提供 Env 变量名（第二个参数）时，它不会自动添加前缀。例 如，如果第二个参数是 ID，Viper 将查找环境变量 ID。 

在使用 Env 变量时，需要注意的一件重要事情是：每次访问该值时都将读取它。Viper 在 调用 BindEnv 时不固定该值。 

还有一个魔法函数 SetEnvKeyReplacer，SetEnvKeyReplacer 允许使用 strings.Replacer 对象来重写 Env 键。

- 如果想在 Get() 调用中使用 - 或者 . ，但希望环境变量使用 _ 分隔符，可以通过 SetEnvKeyReplacer 来实现。

- 比如，设置了环境变量`USER_SECRET_KEY=bVix2WBv0VPfrDrvlLWrhEdzjLpPCNYb`，但想用 `viper.Get("user.secret-key")`，那就调用函数：

- ```go
  viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
  ```

- 上面的代码，在调用 viper.Get() 函数时，会用 _ 替换 . 和 - 。

默认情况下，空环境变量被认为是未设置的，并将返回到下一个配置源。若要将空环境变量视为已设置，可以使用 AllowEmptyEnv 方法。使用环境变量示例如下：

```go
// 使用环境变量
os.Setenv("VIPER_USER_SECRET_ID", "QLdywI2MrmDVjSSv6e95weNRvmteRjfKAuNV")
os.Setenv("VIPER_USER_SECRET_KEY", "bVix2WBv0VPfrDrvlLWrhEdzjLpPCNYb")

viper.AutomaticEnv()                                             // 读取环境变量
viper.SetEnvPrefix("VIPER")                                      // 设置环境变量前缀：VIPER_，如果是viper，将自动转变为大写。
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")) // 将viper.Get(key) key字符串中'.'和'-'替换为'_'
viper.BindEnv("user.secret-key")
viper.BindEnv("user.secret-id", "USER_SECRET_ID") // 绑定环境变量名到key
```

##### 使用标志

6. 使用标志。

Viper 支持 Pflag 包，能够绑定 key 到 Flag。与 BindEnv 类似，在调用绑定方法时，不会设置该值，但在访问它时会设置。

对于单个标志，可以调用 BindPFlag() 进行绑定：

```go
viper.BindPFlag("token", pflag.Lookup("token")) // 绑定单个标志
```

还可以绑定一组现有的 pflags（pflag.FlagSet）：

```go
viper.BindPFlags(pflag.CommandLine) //绑定标志集
```

#### 读取配置 

Viper 提供了如下方法来读取配置：

- Get(key string) interface{} 
- Get(key string)  
- AllSettings() map[string]interface{} 
- IsSet(key string) : bool

每一个 Get 方法在找不到值的时候都会返回零值。为了检查给定的键是否存在，可以使用 IsSet() 方法。可以是 Viper 支持的类型，首字母大写：Bool、Float64、Int、 IntSlice、String、StringMap、StringMapString、StringSlice、Time、Duration。例 如：GetInt()。 

常见的读取配置方法有以下几种。

##### 访问嵌套的键

1. 访问嵌套的键。

例如，加载下面的 JSON 文件：

```json
{
   "host":{
      "address":"localhost",
      "port":5799
   },
   "datastore":{
      "metric":{
         "host":"127.0.0.1",
         "port":3099
      },
      "warehouse":{
         "host":"198.0.0.1",
         "port":2112
      }
   }
}
```

Viper 可以通过**传入 . 分隔的路径来访问嵌套字段**：

```go
viper.GetString("datastore.metric.host") // (返回 "127.0.0.1")
```

如果 datastore.metric 被直接赋值覆盖（被 Flag、环境变量、set() 方法等等），那么 datastore.metric的所有子键都将变为未定义状态，它们被高优先级配置级别覆盖了。 

如果存在与分隔的键路径匹配的键，则**直接返回其值**。例如：

```json
{
   "datastore.metric.host":"0.0.0.0",
   "host":{
      "address":"localhost",
      "port":5799
   },
   "datastore":{
      "metric":{
         "host":"127.0.0.1",
         "port":3099
      },
      "warehouse":{
         "host":"198.0.0.1",
         "port":2112
      }
   }
}
```

通过 viper.GetString 获取值：

```go
viper.GetString("datastore.metric.host") // 返回 "0.0.0.0"
```

##### 反序列化

2. 反序列化。

Viper 可以支持将所有或特定的值解析到结构体、map 等。可以通过两个函数来实现：

- Unmarshal(rawVal interface{}) error 
- UnmarshalKey(key string, rawVal interface{}) error

一个示例：

```go
type config struct {
  Port int
  Name string
  PathMap string `mapstructure:"path_map"`
}
var C config
err := viper.Unmarshal(&C)
if err != nil {
  t.Fatalf("unable to decode into struct, %v", err)
}
```

如果想要解析那些键本身就包含 . (默认的键分隔符）的配置，则需要修改分隔符：

```go
v := viper.NewWithOptions(viper.KeyDelimiter("::"))

v.SetDefault("chart::values", map[string]interface{}{
  "ingress": map[string]interface{}{
    "annotations": map[string]interface{}{
      "traefik.frontend.rule.type": "PathPrefix", 
      "traefik.ingress.kubernetes.io/ssl-redirect": "true",
    },
  },
})

type config struct {
  Chart struct {
    Values map[string]interface{}
  }
}

var C config

v.Unmarshal(&C)
```

Viper 在后台使用 `github.com/mitchellh/mapstructure` 来解析值，其默认情况下使用 mapstructure tags。当需要将 Viper 读取的配置反序列到定义的结构体变 量中时，一定要使用 mapstructure tags。

##### 序列化成字符串

3. 序列化成字符串。

有时候需要将 Viper 中保存的所有设置序列化到一个字符串中，而不是将它们写入到 一个文件中，示例如下：

```go
import (
   yaml "gopkg.in/yaml.v2"
	 // ...
)

func yamlStringSettings() string {
   c := viper.AllSettings()
   bs, err := yaml.Marshal(c)
   if err != nil {
   	 log.Fatalf("unable to marshal config to YAML: %v", err)
   }
   return string(bs)
}
```

### 现代化的命令行框架：Cobra 全解 

Cobra 既是一个可以创建强大的现代 CLI 应用程序的库，也是一个可以生成应用和命令文 件的程序。有许多大型项目都是用 Cobra 来构建应用程序的，例如 Kubernetes、Docker、etcd、Rkt、Hugo 等。 

Cobra 建立在 commands、arguments 和 flags 结构之上。

commands 代表命令， arguments 代表非选项参数，flags 代表选项参数（也叫标志）。

一个好的应用程序应该是易懂的，用户可以清晰地知道如何去使用这个应用程序。应用程序通常遵循如下模式： `APPNAME VERB NOUN --ADJECTIVE` 或者 `APPNAME COMMAND ARG --FLAG`，例如：

```sh
git clone URL --bare # clone 是一个命令，URL是一个非选项参数，bare是一个选项参数
```

这里，VERB 代表动词，NOUN 代表名词，ADJECTIVE 代表形容词。 

Cobra 提供了两种方式来创建命令：Cobra 命令和 Cobra 库。

Cobra 命令可以生成一个 Cobra 命令模板，而命令模板也是通过引用 Cobra 库来构建命令的。

所以，这里直接介 绍如何使用 Cobra 库来创建命令。 

#### 使用 Cobra 库创建命令 

如果要用 Cobra 库编码实现一个应用程序，需要首先创建一个空的 main.go 文件和一个 rootCmd 文件，之后可以根据需要添加其他命令。具体步骤如下：

##### 创建 rootCmd

1. 创建 rootCmd。

```sh
$ mkdir -p newApp2 && cd newApp2
```

通常情况下，会将 rootCmd 放在文件 cmd/root.go 中。

```go
var rootCmd = &cobra.Command{
	Use:   "hugo",
	Short: "Hugo is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

还可以在 init() 函数中定义标志和处理配置，例如 cmd/root.go。

```go
package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	projectBase string
	userLicense string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringVarP(&projectBase, "projectbase", "b", "", "base project directory eg. github.com/spf13/")
	rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "Author name for copyright attribution")
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "Name of license for the project (can provide `licensetext` in config)")
	rootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")
	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.BindPFlag("projectbase", rootCmd.PersistentFlags().Lookup("projectbase"))
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	viper.SetDefault("license", "apache")
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
}
```

##### 创建 main.go

2. 创建 main.go。

还需要一个 main 函数来调用 rootCmd，通常会创建一个 main.go 文件，在 main.go 中调用 rootCmd.Execute() 来执行命令：

```go
package main

import (
	"{pathToYourApp}/cmd"
)

func main() {
	cmd.Execute()
}
```

需要注意，main.go 中不建议放很多代码，通常只需要调用 cmd.Execute() 即可。

##### 添加命令

3. 添加命令。

除了 rootCmd，还可以调用 AddCommand 添加其他命令，通常情况下，会把 其他命令的源码文件放在 cmd/ 目录下。

例如，添加一个 version 命令，可以创建 cmd/version.go 文件，内容为：

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
	},
}
```

本示例中，通过调用 `rootCmd.AddCommand(versionCmd)` 给 rootCmd 命令添加 了一个 versionCmd 命令。

##### 编译并运行

4. 编译并运行。

将 main.go 中{pathToYourApp}替换为对应的路径，例如本示例中 pathToYourApp 为 `github.com/marmotedu/gopractise-demo/cobra/newApp2`。

```sh
$ go mod init github.com/marmotedu/gopractise-demo/cobra/newApp2
# self test: go mod init Code/src/cobra/newApp2
# cd Code/src/cobra/newApp2
$ go build -v .
$ ./newApp2 -h
A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com

Usage:
  hugo [flags]
  hugo [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of Hugo

Flags:
  -a, --author string         Author name for copyright attribution (default "YOUR NAME")
      --config string         config file (default is $HOME/.cobra.yaml)
  -h, --help                  help for hugo
  -l, --license licensetext   Name of license for the project (can provide licensetext in config)
  -b, --projectbase string    base project directory eg. github.com/spf13/
      --viper                 Use Viper for configuration (default true)

Use "hugo [command] --help" for more information about a command.
```

通过步骤一、步骤二、步骤三，就成功创建和添加了 Cobra 应用程序及其命令。 

接下来，再来详细介绍下 Cobra 的核心特性。 

#### 使用标志 

Cobra 可以跟 Pflag 结合使用，实现强大的标志功能。使用步骤如下：

##### 持久化的标志

1. 使用持久化的标志。

标志可以是“持久的”，这意味着该标志可用于它所分配的命令以及该命令下的每个子命令。可以在 rootCmd 上定义持久标志：

```go
rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
```

##### 本地标志

2. 使用本地标志。

也可以分配一个本地标志，本地标志只能在它所绑定的命令上使用：

```go
rootCmd.Flags().StringVarP(&Source, "source", "s", "", "Source directory to read from")
```

--source标志只能在 rootCmd 上引用，而不能在 rootCmd 的子命令上引用。

##### 标志绑定到 Viper

3. 将标志绑定到 Viper。

可以将标志绑定到 Viper，这样就可以使用 viper.Get() 获取标志的值。

```go
var author string

func init() {
  rootCmd.PersistentFlags().StringVar(&author, "author", "YOUR NAME", "Author name for copyright attribution")
  viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
}
```

##### 设置标志为必选

4. 设置标志为必选。

默认情况下，标志是可选的，也可以设置标志为必选，当设置标志为必选，但是没有提供标志时，Cobra 会报错。

```go
rootCmd.Flags().StringVarP(&Region, "region", "r", "", "AWS region (required)")
rootCmd.MarkFlagRequired("region")
```

#### 非选项参数验证 

在命令的过程中，经常会传入非选项参数，并且需要对这些非选项参数进行验证，Cobra 提供了机制来对非选项参数进行验证。

可以使用 Command 的 Args 字段来验证非选项参数。Cobra 也内置了一些验证函数：

- NoArgs：如果存在任何非选项参数，该命令将报错。 
- ArbitraryArgs：该命令将接受任何非选项参数。 
- OnlyValidArgs：如果有任何非选项参数不在 Command 的 ValidArgs 字段中，该命令 将报错。 
- MinimumNArgs(int)：如果没有至少 N 个非选项参数，该命令将报错。 
- MaximumNArgs(int)：如果有多于 N 个非选项参数，该命令将报错。 
- ExactArgs(int)：如果非选项参数个数不为 N，该命令将报错。 
- ExactValidArgs(int)：如果非选项参数的个数不为 N，或者非选项参数不在 Command 的 ValidArgs 字段中，该命令将报错。
- RangeArgs(min, max)：如果非选项参数的个数不在 min 和 max 之间，该命令将报 错。

使用预定义验证函数，示例如下：

```go
var cmd = &cobra.Command{
  Short: "hello",
  Args: cobra.MinimumNArgs(1), // 使用内置的验证函数
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Hello, World!")
  },
}
```

当然也可以自定义验证函数，示例如下：

```go
var cmd = &cobra.Command{
  Short: "hello",
  // Args: cobra.MinimumNArgs(10), // 使用内置的验证函数
  Args: func(cmd *cobra.Command, args []string) error { // 自定义验证函数
    if len(args) < 1 {
      return errors.New("requires at least one arg")
    }
    if myapp.IsValidColor(args[0]) {
      return nil
    }
    return fmt.Errorf("invalid color specified: %s", args[0])
  },
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Hello, World!")
  },
}
```

#### PreRun and PostRun Hooks 

在运行 Run 函数时，可以运行一些钩子函数，比如 PersistentPreRun 和 PreRun 函数在 Run 函数之前执行，PersistentPostRun 和 PostRun 在 Run 函数之后执行。

如果子命令没有指定`Persistent *Run`函数，则子命令将会继承父命令的`Persistent *Run`函 数。这些函数的运行顺序如下：

- PersistentPreRun 
- PreRun 
- Run 
- PostRun 
- PersistentPostRun

注意，父级的 PreRun 只会在父级命令运行时调用，子命令是不会调用的。 

Cobra 还支持很多其他有用的特性，比如：

- 自定义 Help 命令；
- 可以自动添加 --version 标志，输出程序版本信息；
- 当用户提供无效标志或无效命令时，Cobra 可以打印出 usage 信息；
- 当输入的命令有误时，Cobra 会根据注册的命令，推算出可能的命令，等等。 

### 总结 

在开发 Go 项目时，可以通过 Pflag 来解析命令行参数，通过 Viper 来解析配置文 件，用 Cobra 来实现命令行框架。

可以通过 pflag.String()、 pflag.StringP()、 pflag.StringVar()、pflag.StringVarP() 方法来设置命令行参数，并使用 Get来获 取参数的值。

也可以使用 Viper 从命令行参数、环境变量、配置文件等位置读取配置项。最常用的是从配置文件中读取，可以通过 viper.AddConfigPath 来设置配置文件搜索路径，通过 viper.SetConfigFile 和 viper.SetConfigType 来设置配置文件名，通过 viper.ReadInConfig 来读取配置文件。读取完配置文件，然后在程序中使用 Get/Get来读取配置项的值。 

最后，可以使用 Cobra 来构建一个命令行框架，Cobra 可以很好地集成 Pflag 和 Viper。 

### 课后练习

- 研究下 Cobra 的代码，看下 Cobra 是如何跟 Pflag 和 Viper 进行集成的。 
- 思考下，除了 Pflag、Viper、Cobra，在开发过程中还遇到哪些优秀的包，来处理命令行参数、配置文件和启动命令行框架的呢？



## GO 项目之应用构建实战

来聊聊开发应用必须要做的那些事儿。 

应用开发是软件开发工程师最核心的工作。在 7 年的 Go 开发生涯中，构建了大大 小小不下 50 个后端应用，深谙其中的痛点，比如：

- 重复造轮子。同样的功能却每次都要重新开发，浪费非常多的时间和精力不说，每次实 现的代码质量更是参差不齐。 
- 理解成本高。相同的功能，有 N 个服务对应着 N 种不同的实现方式，如果功能升级， 或者有新成员加入，都可能得重新理解 N 次。 
- 功能升级的开发工作量大。一个应用由 N 个服务组成，如果要升级其中的某个功能，需要同时更新 N 个服务的代码。

想要解决上面这些问题，一个比较好的思路是：找出相同的功能，然后用一种优雅的方式去实现它，并通过 Go 包的形式，供所有的服务使用。 

如果面临这些问题，并且正在寻找解决方法，那可以认真学习这一讲。找出服务的通用功能，并给出优雅的构建方式，一劳永逸地解决这些问题。在提高开发效率的同时，也能提高代码质量。 

接下来，先来分析并找出 Go 服务通用的功能。 

### 构建应用的基础：应用的三大基本功能 

目前见到的 Go 后端服务，基本上可以分为 API 服务和非 API 服务两类。

- API 服务：通过对外提供 HTTP/RPC 接口来完成指定的功能。
  - 比如订单服务，通过调用 创建订单的 API 接口，来创建商品订单。 
- 非 API 服务：通过监听、定时运行等方式，而不是通过 API 调用来完成某些任务。
  - 比如数据处理服务，定时从 Redis 中获取数据，处理后存入后端存储中。
  - 再比如消息处理服务，监听消息队列（如 NSQ/Kafka/RabbitMQ），收到消息后进行处理。

对于 API 服务和非 API 服务来说，它们的启动流程基本一致，都可以分为三步：

- 应用框架的构建，这是最基础的一步。 
- 应用初始化。 
- 服务启动。

如下图所示：

![image-20211129220852215](IAM-document.assets/image-20211129220852215.png)

图中，命令行程序、命令行参数解析和配置文件解析，是所有服务都需要具备的功能，这些功能有机结合到一起，共同构成了应用框架。 

所以，要构建的任何一个应用程序，至少要具备命令行程序、命令行参数解析和配置 文件解析这 3 种功能。

- 命令行程序：用来启动一个应用。命令行程序需要实现诸如应用描述、help、参数校验等功能。根据需要，还可以实现命令自动补全、打印命令行参数等高级功能。 
- 命令行参数解析：用来在启动时指定应用程序的命令行参数，以控制应用的行为。 
- 配置文件解析：用来解析不同格式的配置文件。

另外，上述 3 类功能跟业务关系不大，可以抽象成一个统一的框架。应用初始化、创建 API/ 非 API 服务、启动服务，跟业务联系比较紧密，难以抽象成一个统一的框架。 

### iam-apiserver 是如何构建应用框架的？ 

这里，通过讲解 iam-apiserver 的应用构建方式，来讲解下如何构建应用。iam-apiserver 程序的 main 函数位于 app.go 文件中，其构建代码可以简化为：

```go
// Package apiserver does all of the work necessary to create a iam APIServer.
package apiserver

import (
	...
	"github.com/marmotedu/iam/internal/apiserver"
	"github.com/marmotedu/iam/pkg/app"
)

func main() {
  ...
  apiserver.NewApp("iam-apiserver").Run()
}

const commandDesc = `The IAM API server validates and configures data ...`

// NewApp creates a App object with default parameters.
func NewApp(basename string) *app.App {
	opts := options.NewOptions()
	application := app.NewApp("IAM API Server",
		basename,
		app.WithOptions(opts),
		app.WithDescription(commandDesc),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
	)

	return application
}

func run(opts *options.Options) app.RunFunc {
	return func(basename string) error {
		log.Init(opts.Log)
		defer log.Flush()

		cfg, err := config.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}

		return Run(cfg)
	}
}
```

可以看到，是通过调用包 github.com/marmotedu/iam/pkg/app 来构建应用的。 也就是说，将构建应用的功能抽象成了一个 Go 包，通过 Go 包可以提高代码的封装 性和复用性。

iam-authz-server 和 iam-pump 组件也都是通过 github.com/marmotedu/iam/pkg/app 来构建应用的。 

构建应用的流程也很简单，只需要创建一个 application 实例即可：

```go
opts := options.NewOptions()
application := app.NewApp("IAM API Server",
                          basename,
                          app.WithOptions(opts),
                          app.WithDescription(commandDesc),
                          app.WithDefaultValidArgs(),
                          app.WithRunFunc(run(opts)),
                         )
```

在创建应用实例时，传入了下面这些参数。

- IAM API Server：应用的简短描述。 
- basename：应用的二进制文件名。 
- opts：应用的命令行选项。 
- commandDesc：应用的详细描述。 
- run(opts)：应用的启动函数，初始化应用，并最终启动 HTTP 和 GRPC Web 服务。

创建应用时，还可以根据需要来配置应用实例，比如 iam-apiserver 组件在创建应用 时，指定了 WithDefaultValidArgs 来校验命令行非选项参数的默认校验逻辑。 

可以看到，iam-apiserver 通过简单的几行代码，就创建出了一个应用。之所以这么方 便，是因为应用框架的构建代码都封装在了 github.com/marmotedu/iam/pkg/app 包中。

接下来，来重点看下 github.com/marmotedu/iam/pkg/app 包是如何实现的。为了方便描述，在下文中统称为 App 包。 

### App 包设计和实现 

先来看下 App 包目录下的文件：

```sh
[going@dev ~/workspace/golang/src/github.com/marmotedu/iam]$ls pkg/app/
app.go  cmd.go  config.go  doc.go  flag.go  help.go  options.go
```

pkg/app 目录下的 5 个主要文件是 app.go、cmd.go、config.go、flag.go、 options.go，分别实现了应用程序框架中的**应用、命令行程序、命令行参数解析、配置文件解析和命令行选项** 5 个部分，具体关系如下图所示：

![image-20211129222404535](IAM-document.assets/image-20211129222404535.png)

来解释下这张图。应用由命令行程序、命令行参数解析、配置文件解析三部分组成， 命令行参数解析功能通过命令行选项来构建，二者通过接口解耦合：

```go
// CliOptions abstracts configuration options for reading parameters from the
// command line.
type CliOptions interface {
	// AddFlags adds flags to the specified FlagSet object.
	// AddFlags(fs *pflag.FlagSet)
	Flags() (fss cliflag.NamedFlagSets)
	Validate() []error
}
```

通过接口，应用可以定制自己独有的命令行参数。接下来，再来看下如何具体构建应 用的每一部分。 

#### 第 1 步：构建应用 

APP 包提供了 NewApp 函数来创建一个应用：

```go
// NewApp creates a new application instance based on the given application name,
// binary name, and other options.
func NewApp(name string, basename string, opts ...Option) *App {
	a := &App{
		name:     name,
		basename: basename,
	}

	for _, o := range opts {
		o(a)
	}

	a.buildCommand()

	return a
}
```

NewApp 中使用了设计模式中的选项模式，来动态地配置 APP，支持 WithRunFunc、 WithDescription、WithValidArgs 等选项。 

#### 第 2 步：命令行程序构建 

这一步，会使用 Cobra 包来构建应用的命令行程序。 

NewApp 最终会调用 buildCommand 方法来创建 Cobra Command 类型的命令，命令的功能通过指定 Cobra Command 类型的各个字段来实现。

通常可以指定：Use、 Short、Long、SilenceUsage、SilenceErrors、RunE、Args 等字段。 

在 buildCommand 函数中，也会根据应用的设置添加不同的命令行参数，例如：

```go
if !a.noConfig {
  addConfigFlag(a.basename, namedFlagSets.FlagSet("global"))
}
```

上述代码的意思是：如果设置了 noConfig=false，那么就会在命令行参数 global 分组中添加以下命令行选项：

```sh
-c, --config FILE
```

为了更加易用和人性化，命令还具有如下 3 个功能。

- 帮助信息：执行 -h/--help 时，输出的帮助信息。通过 cmd.SetHelpFunc 函数可以 指定帮助信息。
- 使用信息（可选）：当用户提供无效的标志或命令时，向用户显示“使用信息”。通过 cmd.SetUsageFunc 函数，可以指定使用信息。如果不想每次输错命令打印一大堆 usage 信息，可以通过设置 SilenceUsage: true 来关闭掉 usage。 
- 版本信息：打印应用的版本。知道应用的版本号，对故障排查非常有帮助。通过 verflag.AddFlags 可以指定版本信息。例如，App 包通过 github.com/marmotedu/component-base/pkg/version 指定了以下版本信息：

```sh
$ ./iam-apiserver --version
  gitVersion: v0.3.0
  gitCommit: ccc31e292f66e6bad94efb1406b5ced84e64675c
  gitTreeState: dirty
  buildDate: 2020-12-17T12:24:37Z
  goVersion: go1.15.1
  compiler: gc
  platform: linux/amd64
$ ./iam-apiserver --version=raw
  version.Info{GitVersion:"v0.3.0", GitCommit:"ccc31e292f66e6bad94efb1406b5ced84e64675c"
```

接下来，再来看下应用需要实现的另外一个重要功能，也就是命令行参数解析。 

#### 第 3 步：命令行参数解析 

App 包在构建应用和执行应用两个阶段来实现命令行参数解析。 

##### 构建应用阶段

先看构建应用这个阶段。App 包在 buildCommand 方法中通过以下代码段，给应 用添加了命令行参数：

```go
var namedFlagSets cliflag.NamedFlagSets
if a.options != nil {
  namedFlagSets = a.options.Flags()
  fs := cmd.Flags()
  for _, f := range namedFlagSets.FlagSets {
    fs.AddFlagSet(f)
  }
  
  ...
}

if !a.noVersion {
  verflag.AddFlags(namedFlagSets.FlagSet("global"))
}
if !a.noConfig {
  addConfigFlag(a.basename, namedFlagSets.FlagSet("global"))
}
globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
```

namedFlagSets 中引用了 Pflag 包，上述代码先通过 a.options.Flags() 创建并返回 了一批 FlagSet，a.options.Flags() 函数会将 FlagSet 进行分组。

通过一个 for 循 环，将 namedFlagSets 中保存的 FlagSet 添加到 Cobra 应用框架中的 FlagSet 中。 

buildCommand 还会根据应用的配置，选择性添加一些 flag。例如，在 global 分组下添 加 --version 和 --config 选项。 执行 -h 打印命令行参数如下：

```sh
..
Usage:
  iam-apiserver [flags]
  
Generic flags:
  --server.healthz Add self readiness check and install 
  --server.max-ping-count int The max number of ping attempts when server
...

Global flags:
  -h, --help help for iam-apiserver
  --version version[=true] Print version information and quit.
```

这里有两个技巧，可以借鉴。 

- 第一个技巧，将 flag 分组。 

  - 一个大型系统，可能会有很多个 flag，例如 kube-apiserver 就有 200 多个 flag，这时对 flag 分组就很有必要了。

  - 通过分组，可以很快地定位到需要的分组及该分组具有的标志。

  - 例如，想了解 MySQL 有哪些标志，可以找到 MySQL 分组：

  - ```sh
    Mysql flags:
      --mysql.database string
      	Database name for the server to use.
      --mysql.host string
      	MySQL service host address. If left blank, the following related
      --mysql.log-mode int
      	Specify gorm log level. (default 1)
      ...
    ```

- 第二个技巧，flag 的名字带有层级关系。这样不仅可以知道该 flag 属于哪个分组，而且能 够避免重名。例如：

  - ```sh
    $ ./iam-apiserver -h |grep host
      --mysql.host string MySQL service host address.
      --redis.host string Hostname of your Redis server. 
    ```

  - 对于 MySQL 和 Redis， 都可以指定相同的 host 标志，通过 --mysql.host 也可以知道该 flag 隶属于 mysql 分组，代表的是 MySQL 的 host。 

##### 应用执行阶段

再看应用执行阶段。这时会通过 viper.Unmarshal，将配置 Unmarshal 到 Options 变量中。这样就可以使用 Options 变量中的值，来执行后面的业务逻辑。 

传入的 Options 是一个实现了 CliOptions 接口的结构体变量，CliOptions 接口定 义为：

```go
// CliOptions abstracts configuration options for reading parameters from the
// command line.
type CliOptions interface {
	// AddFlags adds flags to the specified FlagSet object.
	// AddFlags(fs *pflag.FlagSet)
	Flags() (fss cliflag.NamedFlagSets)
	Validate() []error
}
```

因为 Options 实现了 Validate 方法，所以就可以在应用框架中调用 Validate 方法来 校验参数是否合法。

另外，还可以通过以下代码，来判断选项是否可补全和打印：如果可以补全，则补全选项；如果可以打印，则打印选项的内容。实现代码如下：

```go
func (a *App) applyOptionRules() error {
	if completeableOptions, ok := a.options.(CompleteableOptions); ok {
		if err := completeableOptions.Complete(); err != nil {
			return err
		}
	}

	if errs := a.options.Validate(); len(errs) != 0 {
		return errors.NewAggregate(errs)
	}

	if printableOptions, ok := a.options.(PrintableOptions); ok && !a.silence {
		log.Infof("%v Config: `%s`", progressMessage, printableOptions.String())
	}

	return nil
}
```

通过配置补全，可以确保一些重要的配置项具有默认值，当这些配置项没有被配置时，程序也仍然能够正常启动。一个大型项目，有很多配置项，不可能对每一个配置项都进行配置。所以，给重要配置项设置默认值，就显得很重要了。 

这里，来看下 iam-apiserver 提供的 Validate 方法：

```go
// internal/apiserver/options/validation.go
// Validate checks Options and return a slice of found errs.
func (o *Options) Validate() []error {
	var errs []error

	errs = append(errs, o.GenericServerRunOptions.Validate()...)
	errs = append(errs, o.GRPCOptions.Validate()...)
	errs = append(errs, o.InsecureServing.Validate()...)
	errs = append(errs, o.SecureServing.Validate()...)
	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.RedisOptions.Validate()...)
	errs = append(errs, o.JwtOptions.Validate()...)
	errs = append(errs, o.Log.Validate()...)
	errs = append(errs, o.FeatureOptions.Validate()...)

	return errs
}
```

可以看到，每个配置分组，都实现了 Validate() 函数，对自己负责的配置进行校验。

通过这种方式，程序会更加清晰。因为只有配置提供者才更清楚如何校验自己的配置项，所以最好的做法是将配置的校验放权给配置提供者（分组）。 

#### 第 4 步：配置文件解析 

在 buildCommand 函数中，通过 addConfigFlag 调用，添加了 -c, --config FILE 命令行参数，用来指定配置文件：

```go
addConfigFlag(a.basename, namedFlagSets.FlagSet("global"))
```

addConfigFlag 函数代码如下：

```go
// addConfigFlag adds flags for a specific server to the specified FlagSet
// object.
func addConfigFlag(basename string, fs *pflag.FlagSet) {
	fs.AddFlag(pflag.Lookup(configFlagName))

	viper.AutomaticEnv()
	viper.SetEnvPrefix(strings.Replace(strings.ToUpper(basename), "-", "_", -1))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.AddConfigPath(".")

			if names := strings.Split(basename, "-"); len(names) > 1 {
				viper.AddConfigPath(filepath.Join(homedir.HomeDir(), "."+names[0]))
			}

			viper.SetConfigName(basename)
		}

		if err := viper.ReadInConfig(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: failed to read configuration file(%s): %v\n", cfgFile, err)
			os.Exit(1)
		}
	})
}
```

addConfigFlag 函数中，指定了 Cobra Command 在执行命令之前，需要做的初始化工 作：

```go
func() {
   if cfgFile != "" {
      viper.SetConfigFile(cfgFile)
   } else {
      viper.AddConfigPath(".")

      if names := strings.Split(basename, "-"); len(names) > 1 {
         viper.AddConfigPath(filepath.Join(homedir.HomeDir(), "."+names[0]))
      }

      viper.SetConfigName(basename)
   }

   if err := viper.ReadInConfig(); err != nil {
      _, _ = fmt.Fprintf(os.Stderr, "Error: failed to read configuration file(%s): %v\n", cfgFile, err)
      os.Exit(1)
   }
}
```

上述代码实现了以下功能：

- 如果命令行参数中没有指定配置文件的路径，则加载默认路径下的配置文件，通过 viper.AddConfigPath、viper.SetConfigName 来设置配置文件搜索路径和配置文件名。通过设置默认的配置文件，可以不用携带任何命令行参数，即可运行程序。 
- 支持环境变量，通过 viper.SetEnvPrefix 来设置环境变量前缀，避免跟系统中的环境变量重名。通过 viper.SetEnvKeyReplacer 重写了 Env 键。

上面，给应用添加了配置文件的命令行参数，并设置在命令执行前，读取配置文件。 

在命令执行时，会将配置文件中的配置项和命令行参数绑定，并将 Viper 的配置 Unmarshal 到传入的 Options 中：

```go
if !a.noConfig {
   if err := viper.BindPFlags(cmd.Flags()); err != nil {
      return err
   }

   if err := viper.Unmarshal(a.options); err != nil {
      return err
   }
}
```

Viper 的配置是命令行参数和配置文件配置 merge 后的配置。如果在配置文件中指定了 MySQL 的 host 配置，并且也同时指定了 --mysql.host 参数，则会优先取命令行参数设置的值。

这里需要注意的是，不同于 YAML 格式的分级方式，配置项是通过点号 . 来分级的。 

至此，已经成功构建了一个优秀的应用框架，接下来看下这个应用框架具有哪些优点吧。 

### 这样构建的应用程序，有哪些优秀特性？ 

借助 Cobra 自带的能力，构建出的应用天然具备帮助信息、使用信息、子命令、子命令自动补全、非选项参数校验、命令别名、PreRun、PostRun 等功能，这些功能对于一个应用来说是非常有用的。 

Cobra 可以集成 Pflag，通过将创建的 Pflag FlagSet 绑定到 Cobra 命令的 FlagSet 中， 使得 Pflag 支持的标志能直接集成到 Cobra 命令中。集成到命令中有很多好处，例如： cobra -h 可以打印出所有设置的 flag，Cobra Command 命令提供的 GenBashCompletion 方法，可以实现命令行选项的自动补全。

通过 viper.BindPFlags 和 viper.ReadInConfig 函数，可以统一配置文件、命令行参数的配置项，使得应用的配置项更加清晰好记。面对不同场景可以选择不同的配置方式，使配置更加灵活。例如：配置 HTTPS 的绑定端口，可以通过 --secure.bind-port 配置， 也可以通过配置文件配置（命令行参数优先于配置文件）：

```yaml
secure:
	bind-address: 0.0.0.0
```

可以通过 viper.GetString("secure.bind-port") 这类方式获取应用的配置，获取 方式更加灵活，而且全局可用。 

将应用框架的构建方法实现成了一个 Go 包，通过 Go 包可以提高应用构建代码的封装性和复用性。 

### 如果想自己构建应用，需要注意些什么？ 

当然，也可以使用其他方式构建应用程序。比如，很多开发者使用如下方式来构建应用：

- 直接在 main.go 文件中通过 gopkg.in/yaml.v3 包解析配置，
- 通过 Go 标准库的 flag 包简单地添加一些命令行参数，例如 --help、--config、--version。 

但是，在自己独立构建应用程序时，很可能会踩这么 3 个坑：

- 构建的应用功能简单，扩展性差，导致后期扩展复杂。 
- 构建的应用没有帮助信息和使用信息，或者信息格式杂乱，增加应用的使用难度。 
- 命令行选项和配置文件支持的配置项相互独立，导致配合应用程序的时候，不知道该使用哪种方式来配置。

对于小的应用，根据需要构建没什么问题，但是对于一个大型项目的话， 还是在应用开发之初，就采用一些功能多、扩展性强的优秀包。

这样，以后随着应用的迭代，可以零成本地进行功能添加和扩展，同时也能体现专业性和技术深度，提高代码质量。 

如果有特殊需求，一定要自己构建应用框架，那么有以下几个建议：

- 应用框架应该清晰易读、扩展性强。
- 应用程序应该至少支持如下命令行选项：-h 打印帮助信息；-v 打印应用程序的版本；- c 支持指定配置文件的路径。 
- 如果应用有很多命令行选项，那么建议支持 --secure.bind-port 这样的长选 项，通过选项名字，就可以知道选项的作用。 
- 配置文件使用 yaml 格式，yaml 格式的配置文件，能支持复杂的配置，还清晰易读。 
- 如果有多个服务，那么要保持所有服务的应用构建方式是一致的。

### 总结 

一个应用框架由命令、命令行参数解析、配置文件解析 3 部分功能组成，可以通过 Cobra 来构建命令，通过 Pflag 来解析命令行参数，通过 Viper 来解析配置文件。

一个项目，可能包含多个应用，这些应用都需要通过 Cobra、Viper、Pflag 来构建。为了不重复造轮子，简化应用的构建，可以将这些功能实现为一个 Go 包，方便直接调用构建应用。 

IAM 项目的应用都是通过 github.com/marmotedu/iam/pkg/app 包来构建的，在构 建时，调用 App 包提供的 NewApp 函数，来构建一个应用：

```go
func NewApp(basename string) *app.App {
	opts := options.NewOptions()
	application := app.NewApp("IAM API Server",
		basename,
		app.WithOptions(opts),
		app.WithDescription(commandDesc),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
	)

	return application
}
```

在构建应用时，只需要提供应用简短 / 详细描述、应用二进制文件名称和命令行选项即可。

App 包会根据 Options 提供的 Flags() 方法，来给应用添加命令行选项。

命令行选项中提供了 -c, --config 选项来指定配置文件，App 包也会加载并解析这个配置文件，并将配置文件和命令行选项相同配置项进行 Merge，最终将配置项的值保存在传入的 Options 变量中，供业务代码使用。 

最后，如果想自己构建应用，给出了一些我的建议：设计一个清晰易读、易扩展的应 用框架；支持一些常见的选项，例如 -h， -v， -c 等；如果应用的命令行选项比较多，建议使用 --secure.bind-port 这样的长选项。 

### 课后练习

- 除了 Cobra、Viper、Pflag 之外，还遇到过哪些比较优秀的包或者工具，可以用来构建应用框架？
- 研究下 iam-apiserver 的命令行选项 Options 是如何通过 Options 的 Flags() 方法 来实现 Flag 分组的，并思考下这样做有什么好处。



## GO项目之 Web 服务

进入实战第三站：服务开发。在这个部分，会讲解 IAM 项目各个服务的构建方式，掌握 Go 开发阶段的各个技能点。 

在 Go 项目开发中，绝大部分情况下，是在写能提供某种功能的后端服务，这些功能 以 RPC API 接口或者 RESTful API 接口的形式对外提供，能提供这两种 API 接口的服务也统称为 Web 服务。

通过介绍 RESTful API 风格的 Web 服务，来介绍下如何实现 Web 服务的核心功能。 

来看下，Web 服务的核心功能有哪些，以及如何开发这些功能。

### Web 服务的核心功能

Web 服务有很多功能，为了便于理解，将这些功能分成了基础功能和高级功能两大 类，并总结在了下面这张图中：

![image-20211130003040332](IAM-document.assets/image-20211130003040332.png)

就按图中的顺序，来串讲下这些功能。 

#### 基础功能

要实现一个 Web 服务，首先要选择通信协议和通信格式。在 Go 项目开发中，有 HTTP+JSON 和 gRPC+Protobuf 两种组合可选。因为 iam-apiserver 主要提供的是 REST 风格的 API 接口，所以选择的是 HTTP+JSON 组合。 

Web 服务最核心的功能是路由匹配。路由匹配其实就是根据(HTTP方法, 请求路径)匹配到处理这个请求的函数，最终由该函数处理这次请求，并返回结果，过程如下图所示：

![image-20211130003226700](IAM-document.assets/image-20211130003226700.png)

一次 HTTP 请求经过路由匹配，最终将请求交由Delete(c *gin.Context)函数来处 理。变量c中存放了这次请求的参数，在 Delete 函数中，可以进行参数解析、参数校 验、逻辑处理，最终返回结果。 

对于大型系统，可能会有很多个 API 接口，API 接口随着需求的更新迭代，可能会有多个版本，为了便于管理，需要对路由进行分组。 

有时候，需要在一个服务进程中，同时开启 HTTP 服务的 80 端口和 HTTPS 的 443 端口，这样就可以做到：对内的服务，访问 80 端口，简化服务访问复杂度；对外的服务，访问更为安全的 HTTPS 服务。显然，没必要为相同功能启动多个服务进程，所以这时候就需要 Web 服务能够支持一进程多服务的功能。 

开发 Web 服务最核心的诉求是：输入一些参数，校验通过后，进行业务逻辑处理，然后返回结果。所以 Web 服务还应该能够进行参数解析、参数校验、逻辑处理、返回结果。 这些都是 Web 服务的业务处理功能。 

上面这些是 Web 服务的基本功能，此外，还需要支持一些高级功能。 

#### 高级功能

在进行 HTTP 请求时，经常需要针对每一次请求都设置一些通用的操作，比如添加 Header、添加 RequestID、统计请求次数等，这就要求 Web 服务能够支持中间件 特性。 

为了保证系统安全，对于每一个请求，都需要进行认证。Web 服务中，通常有两种认 证方式，一种是基于用户名和密码，一种是基于 Token。认证通过之后，就可以继续处理请求了。 

为了方便定位和跟踪某一次请求，需要支持 RequestID，定位和跟踪 RequestID 主要是为了排障。 

最后，当前的软件架构中，很多采用了前后端分离的架构。在前后端分离的架构中，前端访问地址和后端访问地址往往是不同的，浏览器为了安全，会针对这种情况设置跨域请求，所以 Web 服务需要能够处理浏览器的跨域请求。 

到这里，就把 Web 服务的基础功能和高级功能串讲了一遍。当然，上面只介绍了 Web 服务的核心功能，还有很多其他的功能，可以通过学习Gin 的官方文档来了解。 

可以看到，Web 服务有很多核心功能，这些功能可以基于 net/http 包自己封装。 但在实际的项目开发中， 更多会选择使用基于 net/http 包进行封装的优秀开源 Web 框架。

本实战项目选择了 Gin 框架。 接下来，主要看下 Gin 框架是如何实现以上核心功能的，这些功能在实际的开发中可以直接拿来使用。

### 为什么选择 Gin 框架？ 

优秀的 Web 框架有很多，为什么要选择 Gin 呢？在回答这个问题之前，先来看下选择 Web 框架时的关注点。 

在选择 Web 框架时，可以关注如下几点：

- 路由功能； 
- 是否具备 middleware/filter 能力； 
- HTTP 参数（path、query、form、header、body）解析和返回； 
- 性能和稳定性； 
- 使用复杂度； 
- 社区活跃度。

按 GitHub Star 数来排名，当前比较火的 Go Web 框架有 Gin、Beego、Echo、Revel 、 Martini。

经过调研，从中选择了 Gin 框架，原因是 Gin 具有如下特性：

- 轻量级，代码质量高，性能比较高； 
- 项目目前很活跃，并有很多可用的 Middleware； 
- 作为一个 Web 框架，功能齐全，使用起来简单。

接下来，就先详细介绍下 Gin 框架。 

Gin是用 Go 语言编写的 Web 框架，功能完善，使用简单，性能很高。Gin 核心的路由功能是通过一个定制版的 HttpRouter 来实现的，具有很高的路由性能。 

Gin 有很多功能，这里列出了它的一些核心功能：

- 支持 HTTP 方法：GET、POST、PUT、PATCH、DELETE、OPTIONS。 
- 支持不同位置的 HTTP 参数：路径参数（path）、查询字符串参数（query）、表单参数（form）、HTTP 头参数（header）、消息体参数（body）。 
- 支持 HTTP 路由和路由分组。 
- 支持 middleware 和自定义 middleware。 
- 支持自定义 Log。 
- 支持 binding 和 validation，支持自定义 validator。可以 bind 如下参数：query、 path、body、header、form。 
- 支持重定向。 
- 支持 basic auth middleware。 
- 支持自定义 HTTP 配置。 
- 支持优雅关闭。 
- 支持 HTTP2。 
- 支持设置和获取 cookie。

### Gin 是如何支持 Web 服务基础功能的？

先通过一个具体的例子，看下 Gin 是如何支持 Web 服务基础功能的，后面再详细介绍这些功能的用法。 

创建一个 allinone 目录，用来存放示例代码。因为要演示 HTTPS 的用法，所以需要 创建证书文件。具体可以分为两步。 

#### 创建证书

第一步，执行以下命令创建证书：

```sh
cat << 'EOF' > ca.pem
-----BEGIN CERTIFICATE-----
MIICSjCCAbOgAwIBAgIJAJHGGR4dGioHMA0GCSqGSIb3DQEBCwUAMFYxCzAJBgNVBAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxDzANBgNVBAMTBnRlc3RjYTAeFw0xNDExMTEyMjMxMjlaFw0yNDExMDgyMjMxMjlaMFYxCzAJBgNVBAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxDzANBgNVBAMTBnRlc3RjYTCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAwEDfBV5MYdlHVHJ7+L4nxrZy7mBfAVXpOc5vMYztssUI7mL2/iYujiIXM+weZYNTEpLdjyJdu7R5gGUug1jSVK/EPHfc74O7AyZU34PNIP4Sh33N+/A5YexrNgJlPY+E3GdVYi4ldWJjgkAdQah2PH5ACLrIIC6tRka9hcaBlIECAwEAAaMgMB4wDAYDVR0TBAUwAwEB/zAOBgNVHQ8BAf8EBAMCAgQwDQYJKoZIhvcNAQELBQADgYEAHzC7jdYlzAVmddi/gdAeKPausPBG/C2HCWqHzpCUHcKuvMzDVkY/MP2o6JIW2DBbY64bO/FceExhjcykgaYtCH/moIU63+CFOTtR7otyQAWHqXa7q4SbCDlG7DyRFxqG0txPtGvy12lgldA2+RgcigQG
Dfcog5wrJytaQ6UA0wE=
-----END CERTIFICATE-----
EOF

cat << 'EOF' > server.key
-----BEGIN PRIVATE KEY-----
MIICdQIBADANBgkqhkiG9w0BAQEFAASCAl8wggJbAgEAAoGBAOHDFScoLCVJpYDDM4HYtIdV6Ake/sMNaaKdODjDMsux/4tDydlumN+fm+AjPEK5GHhGn1BgzkWF+slf3BxhrA/8dNsnunstVA7ZBgA/5qQxMfGAq4wHNVX77fBZOgp9VlSMVfyd9N8YwbBYAckOeUQadTi2X1S6OgJXgQ0m3MWhAgMBAAECgYAn7qGnM2vbjJNBm0VZCkOkTIWmV10okw7EPJrdL2mkre9NasghNXbE1y5zDshx5Nt3KsazKOxTT8d0Jwh/3KbaN+YYtTCbKGW0pXDRBhwUHRcuRzScjli8Rih5UOCiZkhefUTcRb6xIhZJuQy71tjaSy0pdHZRmYyBYO2YEQ8xoQJBAPrJPhMBkzmEYFtyIEqAxQ/o/A6E+E4w8i+KM7nQCK7qK4JXzyXVAjLfyBZWHGM2uro/fjqPggGD6QH1qXCkI4MCQQDmdKeb2TrKRh5BY1LR81aJGKcJ2XbcDu6wMZK4oqWbTX2KiYn9GB0woM6nSr/Y6iy1u145YzYxEV/iMwffDJULAkB8B2MnyzOg0pNFJqBJuH29bKCcHa8gHJzqXhNO5lAlEbMK95p/P2Wi+4HdaiEIAF1BF326QJcvYKmwSmrORp85AkAlSNxRJ50OWrfMZnBgzVjDx3xG6KsFQVk2ol6VhqL6dFgKUORFUWBvnKSyhjJxurlPEahV6oo6+A+mPhFY8eUvAkAZQyTdupP3XEFQKctGz+9+gKkemDp7LBBMEMBXrGTLPhpEfcjv/7KPdnFHYmhYeBTBnuVmTVWeF98XJ7tIFfJq
-----END PRIVATE KEY-----
EOF

cat << 'EOF' > server.pem
-----BEGIN CERTIFICATE-----
MIICnDCCAgWgAwIBAgIBBzANBgkqhkiG9w0BAQsFADBWMQswCQYDVQQGEwJBVTETMBEGA1UECBMKU29tZS1TdGF0ZTEhMB8GA1UEChMYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMQ8wDQYDVQQDEwZ0ZXN0Y2EwHhcNMTUxMTA0MDIyMDI0WhcNMjUxMTAxMDIyMDI0WjBlMQswCQYDVQQGEwJVUzERMA8GA1UECBMISWxsaW5vaXMxEDAOBgNVBAcTB0NoaWNhZ28xFTATBgNVBAoTDEV4YW1wbGUsIENvLjEaMBgGA1UEAxQRKi50ZXN0Lmdvb2dsZS5jb20wgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAOHDFScoLCVJpYDDM4HYtIdV6Ake/sMNaaKdODjDMsux/4tDydlumN+fm+AjPEK5GHhGn1BgzkWF+slf3BxhrA/8dNsnunstVA7ZBgA/5qQxMfGAq4wHNVX77fBZOgp9VlSMVfyd9N8YwbBYAckOeUQadTi2X1S6OgJXgQ0m3MWhAgMBAAGjazBpMAkGA1UdEwQCMAAwCwYDVR0PBAQDAgXgME8GA1UdEQRIMEaCECoudGVzdC5nb29nbGUuZnKCGHdhdGVyem9vaS50ZXN0Lmdvb2dsZS5iZYISKi50ZXN0LnlvdXR1YmUuY29thwTAqAEDMA0GCSqGSIb3DQEBCwUAA4GBAJFXVifQNub1LUP4JlnX5lXNlo8FxZ2a12AFQs+bzoJ6hM044EDjqyxUqSbVePK0ni3w1fHQB5rY9yYC5f8G7aqqTY1QOhoUk8ZTSTRpnkThy4jjdvTZeLDVBlueZUTDRmy2feY5aZIU18vFDK08dTG0A87pppuv1LNIR3loveU8
-----END CERTIFICATE-----
EOF
```

#### 创建 main.go

第二步，创建 main.go 文件：

```go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"sync"
	"time"
)

type Product struct {
	Username    string    `json:"username" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Category    string    `json:"category" binding:"required"`
	Price       int       `json:"price" binding:"gte=0"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}
type productHandler struct {
	sync.RWMutex
	products map[string]Product
}

func newProductHandler() *productHandler {
	return &productHandler{
		products: make(map[string]Product),
	}
}
func (u *productHandler) Create(c *gin.Context) {
	u.Lock()
	defer u.Unlock()
	// 1. 参数解析
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 2. 参数校验
	if _, ok := u.products[product.Name]; ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("product %s already exist", product.Name)})
		return
	}
	product.CreatedAt = time.Now()
	// 3. 逻辑处理
	u.products[product.Name] = product
	log.Printf("Register product %s success", product.Name)
	// 4. 返回结果
	c.JSON(http.StatusOK, product)
}
func (u *productHandler) Get(c *gin.Context) {
	u.Lock()
	defer u.Unlock()
	product, ok := u.products[c.Param("name")]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Errorf("can not found product %s", c.Param("name"))})
		return
	}
	c.JSON(http.StatusOK, product)
}
func router() http.Handler {
	router := gin.Default()
	productHandler := newProductHandler()
	// 路由分组、中间件、认证
	v1 := router.Group("/v1")
	{
		productv1 := v1.Group("/products")
		{ // 路由匹配
			productv1.POST("", productHandler.Create)
			productv1.GET(":name", productHandler.Get)
		}
	}
	return router
}

func main() {
	var eg errgroup.Group

	// 一进程多端口
	insecureServer := &http.Server{
		Addr:         ":8080",
		Handler:      router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	secureServer := &http.Server{
		Addr:         ":8443",
		Handler:      router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	eg.Go(func() error {
		err := insecureServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		return err
	})
	eg.Go(func() error {
		err := secureServer.ListenAndServeTLS("server.pem", "server.key")
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
```

运行以上代码：

```sh
$ go run main.go
```

打开另外一个终端，请求 HTTP 接口：

```sh
# 创建产品
$ curl -XPOST -H"Content-Type: application/json" -d'{"username":"colin","name":"iphone12","category":"phone","price":8000,"description":"cannot afford"}' http://127.0.0.1:8080/v1/products
{"username":"colin","name":"iphone12","category":"phone","price":8000,"description":"cannot afford","createdAt":"2021-11-30T01:00:18.570392+08:00"}

# 获取产品信息
$ curl -XGET http://127.0.0.1:8080/v1/products/iphone12
{"username":"colin","name":"iphone12","category":"phone","price":8000,"description":"cannot afford","createdAt":"2021-11-30T01:00:18.570392+08:00"}
```

示例代码存放地址为 webfeature。 

另外，Gin 项目仓库中也包含了很多使用示例，如果想详细了解，可以参考 gin examples。 

下面，来详细介绍下 Gin 是如何支持 Web 服务基础功能的。 

#### HTTP/HTTPS 支持 

因为 Gin 是基于 net/http 包封装的一个 Web 框架，所以它天然就支持 HTTP/HTTPS。 

在上述代码中，通过以下方式开启一个 HTTP 服务：

```go
// 一进程多端口
insecureServer := &http.Server{
   Addr:         ":8080",
   Handler:      router(),
   ReadTimeout:  5 * time.Second,
   WriteTimeout: 10 * time.Second,
}
...
err := insecureServer.ListenAndServe()
```

通过以下方式开启一个 HTTPS 服务：

```go
secureServer := &http.Server{
   Addr:         ":8443",
   Handler:      router(),
   ReadTimeout:  5 * time.Second,
   WriteTimeout: 10 * time.Second,
}
...
err := secureServer.ListenAndServeTLS("server.pem", "server.key")
```

#### JSON 数据格式支持 

Gin 支持多种数据通信格式，例如 application/json、application/xml。

可以通过 c.ShouldBindJSON 函数，将 Body 中的 JSON 格式数据解析到指定的 Struct 中，通过 c.JSON函数返回 JSON 格式的数据。 

#### 路由匹配 

Gin 支持两种路由匹配规则。 

第一种匹配规则是**精确匹配**。例如，路由为 /products/:name，匹配情况如下表所示：

![image-20211130011029566](IAM-document.assets/image-20211130011029566.png)

第二种匹配规则是**模糊匹配**。例如，路由为 /products/*name，匹配情况如下表所示：

![image-20211130011101866](IAM-document.assets/image-20211130011101866.png)

#### 路由分组 

Gin 通过 Group 函数实现了路由分组的功能。

路由分组是一个非常常用的功能，可以将相同版本的路由分为一组，也可以将相同 RESTful 资源的路由分为一组。例如：

```go
v1 := router.Group("/v1", gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))
{
   productv1 := v1.Group("/products")
   {
      // 路由匹配
      productv1.POST("", productHandler.Create)
      productv1.GET(":name", productHandler.Get)
   }
   orderv1 := v1.Group("/orders")
   {
      // 路由匹配
      orderv1.POST("", orderHandler.Create)
      orderv1.GET(":name", orderHandler.Get)
   }
}

v2 := router.Group("/v2", gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))
{
   productv2 := v2.Group("/products")
   {
      // 路由匹配
      productv2.POST("", productHandler.Create)
      productv2.GET(":name", productHandler.Get)
   }
}
```

给所有属于 v1 分组的路由都添加 gin.BasicAuth 中间件，以实现认证功能。中间件和认 证，这里先不用深究，下面讲高级功能的时候会介绍到。

#### 一进程多服务 

可以通过以下方式实现一进程多服务：

```go
var eg errgroup.Group

// 一进程多端口
insecureServer := &http.Server{
  ...
}
secureServer := &http.Server{
  ...
}

eg.Go(func() error {
   err := insecureServer.ListenAndServe()
   if err != nil && err != http.ErrServerClosed {
      log.Fatal(err)
   }
   return err
})
eg.Go(func() error {
   err := secureServer.ListenAndServeTLS("server.pem", "server.key")
   if err != nil && err != http.ErrServerClosed {
      log.Fatal(err)
   }
   return err
})

if err := eg.Wait(); err != nil {
   log.Fatal(err)
}
```

上述代码实现了两个相同的服务，分别监听在不同的端口。

这里需要注意的是，为了不阻塞启动第二个服务，需要把 ListenAndServe 函数放在 goroutine 中执行，并且调用 eg.Wait() 来阻塞程序进程，从而让两个 HTTP 服务在 goroutine 中持续监听端口，并提供服务。 

#### 参数解析、参数校验、逻辑处理、返回结果 

此外，Web 服务还应该具有参数解析、参数校验、逻辑处理、返回结果 4 类功能，因为这些功能联系紧密，放在一起来说。 

在 productHandler 的 Create 方法中，通过c.ShouldBindJSON来解析参数，接下 来自己编写校验代码，然后将 product 信息保存在内存中（也就是业务逻辑处理），最后通过c.JSON返回创建的 product 信息。代码如下：

```go
func (u *productHandler) Create(c *gin.Context) {
   u.Lock()
   defer u.Unlock()
   
   // 1. 参数解析
   var product Product
   if err := c.ShouldBindJSON(&product); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
   }
   
   // 2. 参数校验
   if _, ok := u.products[product.Name]; ok {
      c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("product %s already exist", product.Name)})
      return
   }
   product.CreatedAt = time.Now()
   
   // 3. 逻辑处理
   u.products[product.Name] = product
   log.Printf("Register product %s success", product.Name)
   
   // 4. 返回结果
   c.JSON(http.StatusOK, product)
}
```

那这个时候，可能会问：HTTP 的请求参数可以存在不同的位置，Gin 是如何解析的呢？ 这里，先来看下 HTTP 有哪些参数类型。HTTP 具有以下 5 种参数类型：

- 路径参数（path）。例如`gin.Default().GET("/user/:name", nil)`， name 就是路径参数。 
- 查询字符串参数（query）。例如`/welcome? firstname=Lingfei&lastname=Kong`，firstname 和 lastname 就是查询字符串参数。 
- 表单参数（form）。例如`curl -X POST -F 'username=colin' -F 'password=colin1234' http://mydomain.com/login`，username 和 password 就是表单参数。 
- HTTP 头参数（header）。例如`curl -X POST -H 'Content-Type: application/json' -d '{"username":"colin","password":"colin1234"}' http://mydomain.com/login`，Content-Type 就是 HTTP 头参数。 
- 消息体参数（body）。例如`curl -X POST -H 'Content-Type: application/json' -d '{"username":"colin","password":"colin1234"}' http://mydomain.com/login`，username 和 password 就是消息体参数。

Gin 提供了一些函数，来分别读取这些 HTTP 参数，每种类别会提供两种函数，一种函数可以直接读取某个参数的值，另外一种函数会把同类 HTTP 参数绑定到一个 Go 结构体中。

比如，有如下路径参数：

```go
gin.Default().GET("/:name/:id", nil)
```

可以直接读取每个参数：

```GO
name := c.Param("name")
action := c.Param("action")
```

也可以将所有的路径参数，绑定到结构体中：

```GO
type Person struct {
   ID string `uri:"id" binding:"required,uuid"`
   Name string `uri:"name" binding:"required"`
}

if err := c.ShouldBindUri(&person); err != nil {
   // normal code
   return
}
```

Gin 在绑定参数时，是通过结构体的 tag 来判断要绑定哪类参数到结构体中的。这里要注意，不同的 HTTP 参数有不同的结构体 tag。

- 路径参数：uri。 
- 查询字符串参数：form。 
- 表单参数：form。 
- HTTP 头参数：header。 
- 消息体参数：会根据 Content-Type，自动选择使用 json 或者 xml，也可以调用 ShouldBindJSON 或者 ShouldBindXML 直接指定使用哪个 tag。

针对每种参数类型，Gin 都有对应的函数来获取和绑定这些参数。这些函数都是基于如下两个函数进行封装的：

1. **ShouldBindWith**(obj interface{}, b binding.Binding) error

非常重要的一个函数，很多 ShouldBindXXX 函数底层都是调用 ShouldBindWith 函数来完成参数绑定的。

该函数会根据传入的绑定引擎，将参数绑定到传入的结构体指针中，如果绑定失败，只返回错误内容，但不终止 HTTP 请求。

ShouldBindWith 支持多种绑定引擎，例如 binding.JSON、binding.Query、binding.Uri、binding.Header 等，更详细的信息可以参考 binding.go(pkg/mod/github.com/gin-gonic/gin@v1.6.3/binding/binding.go)。

2.  **MustBindWith**(obj interface{}, b binding.Binding) error

这是另一个非常重要的函数，很多 BindXXX 函数底层都是调用 MustBindWith 函数来完成参数绑定的。该函数会根据传入的绑定引擎，将参数绑定到传入的结构体指针中，如果绑定失败，返回错误并终止请求，返回 HTTP 400 错误。

MustBindWith 所支持的绑定引擎跟 ShouldBindWith 函数一样。 

Gin 基于 ShouldBindWith 和 MustBindWith 这两个函数，又衍生出很多新的 Bind 函数。这些函数可以满足不同场景下获取 HTTP 参数的需求。

Gin 提供的函数可以获取 5 个类别的 HTTP 参数。

- 路径参数：ShouldBindUri、BindUri； 
- 查询字符串参数：ShouldBindQuery、BindQuery； 
- 表单参数：ShouldBind； 
- HTTP 头参数：ShouldBindHeader、BindHeader； 
- 消息体参数：ShouldBindJSON、BindJSON 等。

每个类别的 Bind 函数，详细信息可以参考 Gin 提供的 Bind 函数。 

这里要注意，Gin 并没有提供类似 ShouldBindForm、BindForm 这类函数来绑定表单参 数，但可以通过 ShouldBind 来绑定表单参数。

当 HTTP 方法为 GET 时， ShouldBind 只绑定 Query 类型的参数；当 HTTP 方法为 POST 时，会先检查 content-type 是否是 json 或者 xml，如果不是，则绑定 Form 类型的参数。 

所以，ShouldBind 可以绑定 Form 类型的参数，但前提是 HTTP 方法是 POST，并且 content-type 不是 application/json、application/xml。 

在 Go 项目开发中，建议使用 ShouldBindXXX，这样可以确保设置的 HTTP Chain（Chain 可以理解为一个 HTTP 请求的一系列处理插件）能够继续被执行。

### Gin 是如何支持 Web 服务高级功能的？ 

上面介绍了 Web 服务的基础功能，这里再来介绍下高级功能。

Web 服务可以具备多个高级功能，但比较核心的高级功能是中间件、认证、RequestID、跨域和优雅关停。 

#### 中间件 

Gin 支持中间件，HTTP 请求在转发到实际的处理函数之前，会被一系列加载的中间件进行处理。

在中间件中，可以解析 HTTP 请求做一些逻辑处理，例如：跨域处理或者生成 XRequest-ID 并保存在 context 中，以便追踪某个请求。处理完之后，可以选择中断并返回这次请求，也可以选择将请求继续转交给下一个中间件处理。当所有的中间件都处理完之后，请求才会转给路由函数进行处理。

具体流程如下图：

![image-20211130014948438](IAM-document.assets/image-20211130014948438.png)

通过中间件，可以实现对所有请求都做统一的处理，提高开发效率，并使代码更简 洁。但是，因为所有的请求都需要经过中间件的处理，可能会增加请求延时。对于中间件特性，有如下建议：

- 中间件做成可加载的，通过配置文件指定程序启动时加载哪些中间件。
- 只将一些通用的、必要的功能做成中间件。 
- 在编写中间件时，一定要保证中间件的代码质量和性能。

##### 默认中间件

在 Gin 中，可以通过 gin.Engine 的 Use 方法来加载中间件。中间件可以加载到不同的位置上，而且不同的位置作用范围也不同，例如：

```go
router := gin.New()

router.Use(gin.Logger(), gin.Recovery()) // 中间件作用于所有的HTTP请求

v1 := router.Group("/v1").Use(gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))
v1.POST("/login", Login).Use(gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))
```

Gin 框架**本身支持的一些中间件**。

- gin.Logger()：Logger 中间件会将日志写到 gin.DefaultWriter，gin.DefaultWriter 默认为 os.Stdout。 
- gin.Recovery()：Recovery 中间件可以从任何 panic 恢复，并且写入一个 500 状态码。 
- gin.CustomRecovery(handle gin.RecoveryFunc)：类似 Recovery 中间件，但是 在恢复时还会调用传入的 handle 方法进行处理。 
- gin.BasicAuth()：HTTP 请求基本认证（使用用户名和密码进行认证）。

##### 自定义中间件

另外，Gin 还支持**自定义中间件**。中间件其实是一个函数，函数类型为 gin.HandlerFunc，HandlerFunc 底层类型为 func(*Context)。如下是一个 Logger 中间 件的实现：

```go
package main

import (
   "github.com/gin-gonic/gin"
   "log"
   "time"
)

func Logger() gin.HandlerFunc {
   return func(c *gin.Context) {
      t := time.Now()
      // 设置变量example
      c.Set("example", "12345")
      // 请求之前
      c.Next()
      // 请求之后
      latency := time.Since(t)
      log.Print(latency)
      // 访问我们发送的状态
      status := c.Writer.Status()
      log.Println(status)
   }
}

func main() {
   r := gin.New()
   r.Use(Logger())
  // $ curl -XGET http://127.0.0.1:8080/test
   r.GET("/test", func(c *gin.Context) {
      example := c.MustGet("example").(string)
      // it would print: "12345"
      log.Println(example)
   })
   // Listen and serve on 0.0.0.0:8080
   r.Run(":8080")
}
```

另外，还有很多开源的中间件可供选择，把一些常用的总结在了表格里：

![image-20211130020637444](IAM-document.assets/image-20211130020637444.png)

#### 认证、RequestID、跨域 

认证、RequestID、跨域这三个高级功能，都可以通过 Gin 的中间件来实现，例如：

```go
package auth

import (
   "github.com/gin-gonic/gin"
   "time"
)

func main() {
   router := gin.New()
   // 认证
   router.Use(gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))
   
   // RequestID
   router.Use(requestid.New(requestid.Config{
      Generator: func() string {
         return "test"
      },
   }))

   // 跨域
   // CORS for https://foo.com and https://github.com origins, allowing:
   // - PUT and PATCH methods
   // - Origin header
   // - Credentials share
   // - Preflight requests cached for 12 hours
   router.Use(cors.New(cors.Config{
      AllowOrigins:     []string{"https://foo.com"},
      AllowMethods:     []string{"PUT", "PATCH"},
      AllowHeaders:     []string{"Origin"},
      ExposeHeaders:    []string{"Content-Length"},
      AllowCredentials: true,
      AllowOriginFunc: func(origin string) bool {
         return origin == "https://github.com"
      },
      MaxAge: 12 * time.Hour,
   }))
}
```

#### 优雅关停 

Go 项目上线后，还需要不断迭代来丰富项目功能、修复 Bug 等，这也就意味着，要不断地重启 Go 服务。

对于 HTTP 服务来说，如果访问量大，重启服务的时候可能还有很多连接没有断开，请求没有完成。如果这时候直接关闭服务，这些连接会直接断掉， 请求异常终止，这就会对用户体验和产品口碑造成很大影响。

因此，这种关闭方式不是一种优雅的关闭方式。 

这时候，期望 HTTP 服务可以在处理完所有请求后，正常地关闭这些连接，也就是优 雅地关闭服务。

有两种方法来优雅关闭 HTTP 服务，分别是借助第三方的 Go 包和自己编码实现。

##### 第三方 Go 包 

方法一：借助第三方的 Go 包 

如果使用第三方的 Go 包来实现优雅关闭，目前用得比较多的包是 fvbock/endless。可以使用 fvbock/endless 来替换掉 net/http 的 ListenAndServe 方法，例如：

```go
router := gin.Default()
router.GET("/", handler)
// [...]
endless.ListenAndServe(":4242", router)
```

##### 编码实现

方法二：编码实现

借助第三方包的好处是可以稍微减少一些编码工作量，但缺点是引入了一个新的依赖包， 因此更倾向于自己编码实现。

Go 1.8 版本或者更新的版本，http.Server 内置的 Shutdown 方法，已经实现了优雅关闭。下面是一个示例：

```go
package main

import (
   "context"
   "github.com/gin-gonic/gin"
   "log"
   "net/http"
   "os"
   "os/signal"
   "syscall"
   "time"
)

func main() {
   router := gin.Default()

   router.GET("/", func(c *gin.Context) {
      time.Sleep(5 * time.Second)
      c.String(http.StatusOK, "Welcome Gin Server")
   })

   srv := &http.Server{
      Addr:    ":8080",
      Handler: router,
   }

   // Initializing the server in a goroutine so that
   // it won't block the graceful shutdown handling below
   go func() {
      if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
         log.Fatalf("listen: %s\n", err)
      }
   }()

   // Wait for interrupt signal to gracefully shutdown the server with
   // a timeout of 5 seconds.
   quit := make(chan os.Signal)
   // kill (no param) default send syscall.SIGTERM
   // kill -2 is syscall.SIGINT
   // kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
   signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
   <-quit
   log.Println("Shutting down server...")

   // The context is used to inform the server it has 5 seconds to finish
   // the request it is currently handling
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   if err := srv.Shutdown(ctx); err != nil {
      log.Fatal("Server forced to shutdown:", err)
   }
   log.Println("Server exiting")
}
```

上面的示例中，需要把 srv.ListenAndServe 放在 goroutine 中执行，这样才不会阻塞到 srv.Shutdown 函数。因为把 srv.ListenAndServe 放在了 goroutine 中，所以需要一种可以让整个进程常驻的机制。 

这里，借助了无缓冲 channel，并且调用 signal.Notify 函数将该 channel 绑定到 SIGINT、SIGTERM 信号上。这样，收到 SIGINT、SIGTERM 信号后，quilt 通道会被写入值，从而结束阻塞状态，程序继续运行，执行 srv.Shutdown(ctx)，优雅关停 HTTP 服务。 

### 总结 

主要学习了 Web 服务的核心功能，以及如何开发这些功能。在实际的项目开发中， 一般会使用基于 net/http 包进行封装的优秀开源 Web 框架。 

当前比较火的 Go Web 框架有 Gin、Beego、Echo、Revel、Martini。可以根据需要进行选择。比较推荐 Gin，Gin 也是目前比较受欢迎的 Web 框架。

Gin Web 框架支持 Web 服务的很多基础功能，例如 HTTP/HTTPS、JSON 格式的数据、路由分组和匹配、一 进程多服务等。 

另外，Gin 还支持 Web 服务的一些高级功能，例如 中间件、认证、RequestID、跨域和 优雅关停等。 

### 课后练习

- 使用 Gin 框架编写一个简单的 Web 服务，要求该 Web 服务可以解析参数、校验参 数，并进行一些简单的业务逻辑处理，最终返回处理结果。
- 思考下，如何给 iam-apiserver 的 /healthz 接口添加一个限流中间件，用来限制请求 /healthz 的频率。



## GO项目之访问认证

来聊聊如何进行访问认证。 

保证应用的安全是软件开发的最基本要求，有多种途径来保障应用的安全，例如网络 隔离、设置防火墙、设置 IP 黑白名单等。这些更多是从运维角度来解决应用的安全问题。作为开发者，也可以从软件层面来保证应用的安全，这可以通过认证 来实现。 

这一讲，以 HTTP 服务为例，来介绍下当前常见的四种认证方法：Basic、Digest、 OAuth、Bearer。

IAM 项目使用了 Basic、Bearer 两种认证方法。这一讲，先来介绍下这四种认证方法， 下一讲，会介绍下 IAM 项目是如何设计和实现访问认证功能的。

### 认证和授权有什么区别？ 

在介绍四种基本的认证方法之前，先区分下认证和授权，这是很多开发者都容易 搞混的两个概念。

- 认证（Authentication，英文缩写 authn）：用来验证某个用户是否具有访问系统的权限。如果认证通过，该用户就可以访问系统，从而创建、修改、删除、查询平台支持的资源。 
- 授权（Authorization，英文缩写 authz）：用来验证某个用户是否具有访问某个资源的权限，如果授权通过，该用户就能对资源做增删改查等操作。

通过下面的图片，来明白二者的区别：

![image-20211201002856010](IAM-document.assets/image-20211201002856010.png)

图中，有一个仓库系统，用户 james、colin、aaron 分别创建了 Product-A、 Product-B、Product-C。

现在用户 colin 通过用户名和密码（认证）成功登陆到仓库系统 中，但他尝试访问 Product-A、Product-C 失败，因为这两个产品不属于他（授权失败），但他可以成功访问自己创建的资源 Product-B（授权成功）。

由此可见：**认证证明了你是谁，授权决定了你能做什么**。 

上面，介绍了认证和授权的区别。那么接下来，就回到这一讲的重心：应用程序如何进行访问认证。 

### 四种基本的认证方式 

常见的认证方式有四种，分别是 Basic、Digest、OAuth 和 Bearer。先来看下 Basic 认 证。 

#### Basic

Basic 认证（基础认证），是最简单的认证方式。它简单地将用户名:密码进行 base64 码后，放到 HTTP Authorization Header 中。

HTTP 请求到达后端服务后，后端服务会解析出 Authorization Header 中的 base64 字符串，解码获取用户名和密码，并将用户名和密码跟数据库中记录的值进行比较，如果匹配则认证通过。例如：

```go
// base64 编码
$ basic=`echo -n 'admin:Admin@2021'|base64`
// base64 反编码
// $ echo YWRtaW46QWRtaW5AMjAyMQ==|base64 -d

$ curl -XPOST -H"Authorization: Basic ${basic}" http://127.0.0.1:8080/login
{"expire":"2021-12-02T00:35:33+08:00","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2MzgzNzY1MzMsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2MzgyOTAxMzMsInN1YiI6ImFkbWluIn0.1Pg1ynCuzWo_HgLzwmwDuB2xP4Vm620xjD_fSo-ffkI"}
```

通过 base64 编码，可以将密码以非明文的方式传输，增加一定的安全性。但是，base64 不是加密技术，入侵者仍然可以截获 base64 字符串，并反编码获取用户名和密码。

另外，即使 Basic 认证中密码被加密，入侵者仍可通过加密后的用户名和密码进行**重放攻击**。 

所以，Basic 认证虽然简单，但极不安全。使用 Basic 认证的唯一方式就是将它和 SSL 配合使用，来确保整个认证过程是安全的。 

IAM 项目中，为了支持前端通过用户名和密码登录，仍然使用了 Basic 认证，但前后端使用 HTTPS 来通信，保证了认证的安全性。 

这里需要注意，在设计系统时，要遵循一个通用的原则：不要在请求参数中使用明文密码，也不要在任何存储中保存明文密码。

#### Digest 

Digest 认证（摘要认证），是另一种 HTTP 认证协议，它与基本认证兼容，但修复了基本认证的严重缺陷。

Digest 具有如下特点：

- 绝不会用明文方式在网络上发送密码。 
- 可以有效防止恶意用户进行重放攻击。 
- 可以有选择地防止对报文内容的篡改。

摘要认证的过程见下图：

![image-20211201004512855](IAM-document.assets/image-20211201004512855.png)

在上图中，完成摘要认证需要下面这四步：

- 客户端请求服务端的资源。
- 在客户端能够证明它知道密码从而确认其身份之前，服务端认证失败，返回401 Unauthorized，并返回WWW-Authentication头，里面包含认证需要的信息。 
- 客户端根据WWW-Authentication头中的信息，选择加密算法，并使用密码随机数 nonce，计算出密码摘要 response，并再次请求服务端。 
- 服务器将客户端提供的密码摘要与服务器内部计算出的摘要进行对比。如果匹配，就说明客户端知道密码，认证通过，并返回一些与授权会话相关的附加信息，放在 Authorization-Info 中。

WWW-Authentication头中包含的信息见下表：

![image-20211201004846609](IAM-document.assets/image-20211201004846609.png)

虽然使用摘要可以避免密码以明文方式发送，一定程度上保护了密码的安全性，但是仅仅 隐藏密码并不能保证请求是安全的。因为请求（包括密码摘要）仍然可以被截获，这样就 可以重放给服务器，带来安全问题。 

为了防止**重放攻击**，服务器向客户端发送了密码随机数 nonce，nonce 每次请求都会变化。客户端会根据 nonce 生成密码摘要，这种方式，可以使摘要随着随机数的变化而变化。服务端收到的密码摘要只对特定的随机数有效，而没有密码的话，攻击者就无法计算 出正确的摘要，这样就可以防止重放攻击。

摘要认证可以保护密码，比基本认证安全很多。但摘要认证并不能保护内容，所以仍然要与 HTTPS 配合使用，来确保通信的安全。 

#### OAuth 

OAuth（开放授权）是一个开放的授权标准，允许用户让第三方应用访问该用户在某一 Web 服务上存储的私密资源（例如照片、视频、音频等），而无需将用户名和密码提供给 第三方应用。

OAuth 目前的版本是 2.0 版。 

OAuth2.0 一共分为四种授权方式，分别为密码式、隐藏式、拼接式和授权码模式。接下来，就具体介绍下每一种授权方式。 

##### 密码式

第一种，密码式。

密码式的授权方式，就是用户把用户名和密码直接告诉给第三方应用， 然后第三方应用使用用户名和密码换取令牌。所以，使用此授权方式的前提是无法采用其 他授权方式，并且用户高度信任某应用。 

认证流程如下：

- 网站 A 向用户发出获取用户名和密码的请求； 
- 用户同意后，网站 A 凭借用户名和密码向网站 B 换取令牌； 
- 网站 B 验证用户身份后，给出网站 A 令牌，网站 A 凭借令牌可以访问网站 B 对应权限的资源。

##### 隐藏式

第二种，隐藏式。

这种方式适用于前端应用。认证流程如下：

- A 网站提供一个跳转到 B 网站的链接，用户点击后跳转至 B 网站，并向用户请求授权； 
- 用户登录 B 网站，同意授权后，跳转回 A 网站指定的重定向 redirect_url 地址，并携带 B 网站返回的令牌，用户在 B 网站的数据给 A 网站使用。

这个授权方式存在着“中间人攻击”的风险，因此只能用于一些安全性要求不高的场景， 并且令牌的有效时间要非常短。 

##### 凭借式

第三种，凭借式。

这种方式是在命令行中请求授权，适用于没有前端的命令行应用。认证 流程如下：

- 应用 A 在命令行向应用 B 请求授权，此时应用 A 需要携带应用 B 提前颁发的 secretID 和 secretKey，其中 secretKey 出于安全性考虑，需在后端发送；
- 应用 B 接收到 secretID 和 secretKey，并进行身份验证，验证通过后返回给应用 A 令 牌。

##### 授权码式

第四种，授权码模式。

这种方式就是第三方应用先提前申请一个授权码，然后再使用授权 码来获取令牌。相对来说，这种方式安全性更高，前端传送授权码，后端存储令牌，与资 源的通信都是在后端，可以避免令牌的泄露导致的安全问题。

认证流程如下：

![image-20211201010014316](IAM-document.assets/image-20211201010014316.png)

- A 网站提供一个跳转到 B 网站的链接 +redirect_url，用户点击后跳转至 B 网站； 
- 用户携带向 B 网站提前申请的 client_id，向 B 网站发起身份验证请求； 
- 用户登录 B 网站，通过验证，授予 A 网站权限，此时网站跳转回 redirect_url，其中会 有 B 网站通过验证后的授权码附在该 url 后； 
- 网站 A 携带授权码向网站 B 请求令牌，网站 B 验证授权码后，返回令牌即 access_token。

#### Bearer 

Bearer 认证，也称为令牌认证，是一种 HTTP 身份验证方法。

Bearer 认证的核心是 bearer token。bearer token 是一个加密字符串，通常由服务端根据密钥生成。客户端在请求服务端时，必须在请求头中包含`Authorization: Bearer <token>`。服务端收到请求后，解析出 `<token>` ，并校验 `<token>`的合法性，如果校验通过，则认证通过。

跟基本认证一样，Bearer 认证需要配合 HTTPS 一起使用，来保证认证安全性。 当前最流行的 token 编码方式是 JSON Web Token（JWT，音同 jot，详见 JWT RFC 7519）。

接下来，通过讲解 JWT 认证来帮助了解 Bearer 认证的原理。 

### 基于 JWT 的 Token 认证机制实现 

在典型业务场景中，为了区分用户和保证安全，必须对 API 请求进行鉴权，但是不能要求 每一个请求都进行登录操作。

合理做法是，在第一次登录之后产生一个有一定有效期的 token，并将它存储在浏览器的 Cookie 或 LocalStorage 之中。之后的请求都携带这个 token ，请求到达服务器端后，服务器端用这个 token 对请求进行认证。

在第一次登录之 后，服务器会将这个 token 用文件、数据库或缓存服务器等方法存下来，用于之后请求中 的比对。 

或者也可以采用更简单的方法：直接用密钥来签发 Token。这样，就可以省下额外的存 储，也可以减少每一次请求时对数据库的查询压力。这种方法在业界已经有一种标准的实 现方式，就是 JWT。 

#### JWT 简介 

JWT 是 Bearer Token 的一个具体实现，由 JSON 数据格式组成，通过 HASH 散列算法生 成一个字符串。该字符串可以用来进行授权和信息交换。 

使用 JWT Token 进行认证有很多优点，比如说无需在服务端存储用户数据，可以减轻服务 端压力；而且采用 JSON 数据格式，比较易读。除此之外，使用 JWT Token 还有跨语 言、轻量级等优点。

#### JWT 认证流程 

使用 JWT Token 进行认证的流程如下图：

![image-20211201010910790](IAM-document.assets/image-20211201010910790.png)

具体可以分为四步：

- 客户端使用用户名和密码请求登录。 
- 服务端收到请求后，会去验证用户名和密码。如果用户名和密码跟数据库记录不一致， 则验证失败；如果一致则验证通过，服务端会签发一个 Token 返回给客户端。 
- 客户端收到请求后会将 Token 缓存起来，比如放在浏览器 Cookie 中或者 LocalStorage 中，之后每次请求都会携带该 Token。 
- 服务端收到请求后，会验证请求中的 Token，验证通过则进行业务逻辑处理，处理完后 返回处理后的结果。

#### JWT 格式 

JWT 由三部分组成，分别是 Header、Payload 和 Signature，它们之间用圆点.连接，并使用 Base64 编码，例如：

```sh
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJpYW0uYXV0aHoubWFybW90ZWR1LmNvbSIsImV4cCI6MTYwNDE1ODk4NywiaWF0IjoxNjA0MTUxNzg3LCJpc3MiOiJpYW1jdGwiLCJuYmYiOjE2MDQxNTE3ODd9.LjxrK9DuAwAzUD8-9v43NzWBN7HXsSLfebw92DKd1JQ
```

JWT 中，每部分包含的信息见下图：

![image-20211201011208014](IAM-document.assets/image-20211201011208014.png)

下面来具体介绍下这三部分，以及它们包含的信息。

##### Header

1. Header

JWT Token 的 Header 中，包含两部分信息：一是 Token 的类型，二是 Token 所使用的 加密算法。 例如：

```json
{
  "typ": "JWT",
  "alg": "HS256"
}
```

参数说明：

- typ：说明 Token 类型是 JWT。 
- alg：说明 Token 的加密算法，这里是 HS256（alg 算法可以有多种）。

这里，将 Header 进行 base64 编码：

```sh
$ echo -n '{"typ":"JWT","alg":"HS256"}'|base64
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9
```

在某些场景下，可能还会有 kid 选项，用来标识一个密钥 ID，例如：

```json
{
  "alg": "HS256",
  "kid": "XhbY3aCrfjdYcP1OFJRu9xcno8JzSbUIvGE2",
  "typ": "JWT"
}
```

#####  Payload

2. Payload（载荷）

Payload 中携带 Token 的具体内容由三部分组成：JWT 标准中注册的声明（可选）、公 共的声明、私有的声明。下面来分别看下。 

JWT 标准中注册的声明部分，有以下标准字段：

![image-20211201011944840](IAM-document.assets/image-20211201011944840.png)

本例中的 payload 内容为：

```json
{
  "aud": "iam.authz.marmotedu.com",
  "exp": 1604158987,
  "iat": 1604151787,
  "iss": "iamctl",
  "nbf": 1604151787
}
```

这里，将 Payload 进行 base64 编码：

```sh
$ echo -n '{"aud":"iam.authz.marmotedu.com","exp":1604158987,"iat":1604151787,"iss":"iamctl","nbf":1604151787}' |base64
eyJhdWQiOiJpYW0uYXV0aHoubWFybW90ZWR1LmNvbSIsImV4cCI6MTYwNDE1ODk4NywiaWF0IjoxNjA0MTUxNzg3LCJpc3MiOiJpYW1jdGwiLCJuYmYiOjE2MDQxNTE3ODd9
```

除此之外，还有公共的声明和私有的声明。

- 公共的声明可以添加任何的需要的信息，一般 添加用户的相关信息或其他业务需要的信息，注意不要添加敏感信息；
- 私有声明是客户端和服务端所共同定义的声明，因为 base64 是对称解密的，所以一般不建议存放敏感信 息。

##### Signature

3. Signature（签名）

Signature 是 Token 的签名部分，通过如下方式生成：将 Header 和 Payload 分别 base64 编码后，用 . 连接。

然后再使用 Header 中声明的加密方式，利用 secretKey 对连接后的字符串进行加密，加密后的字符串即为最终的 Token。 

secretKey 是密钥，保存在服务器中，一般通过配置文件来保存，例如：

![image-20211201012500708](IAM-document.assets/image-20211201012500708.png)

这里要注意，密钥一定不能泄露。密钥泄露后，入侵者可以使用该密钥来签发 JWT Token，从而入侵系统。 

最后生成的 Token 如下：

```sh
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJpYW0uYXV0aHoubWFybW90ZWR1LmNvbSIsImV4cCI6MTYwNDE1ODk4NywiaWF0IjoxNjA0MTUxNzg3LCJpc3MiOiJpYW1jdGwiLCJuYmYiOjE2MDQxNTE3ODd9.LjxrK9DuAwAzUD8-9v43NzWBN7HXsSLfebw92DKd1JQ
```

签名后服务端会返回生成的 Token，客户端下次请求会携带该 Token。服务端收到 Token 后会解析出 header.payload，然后用相同的加密算法和密钥，对 header.payload 再进行一次加密得到 Signature，并对比加密后的 Signature 和收到的 Signature 是否相同。如果相同则验证通过，不相同则返回HTTP 401 Unauthorized的错误。 

最后，关于 JWT 的使用，还有两点建议：

- 不要存放敏感信息在 Token 里； 
- Payload 中的 exp 值不要设置的太大，一般线上版是2h，开发版随意。当然， 也可以根据需要自行设置。

### 总结 

在开发 Go 应用时，需要通过认证来保障应用的安全。认证，用来验证某个用户是否 具有访问系统的权限，如果认证通过，该用户就可以访问系统，从而创建、修改、删除、 查询平台支持的资源。

业界目前有四种常用的认证方式：Basic、Digest、OAuth、 Bearer。其中 Basic 和 Bearer 用得最多。 

Basic 认证通过用户名和密码来进行认证，主要用在用户登录场景；Bearer 认证通过 Token 来进行认证，通常用在 API 调用场景。不管是 Basic 认证还是 Bearer 认证，都需 要结合 HTTPS 来使用，来最大程度地保证请求的**安全性**。 

Basic 认证简单易懂，但是 Bearer 认证有一定的复杂度，所以这一讲的后半部分通过 JWT Token，讲解了 Bearer Token 认证的原理。 

JWT Token 是 Bearer 认证的一种比较好的实现，主要包含了 3 个部分：

- Header：包含了 Token 的类型、Token 使用的加密算法。在某些场景下，还可以添加 kid 字段，用来标识一个密钥 ID。 
- Payload：Payload 中携带 Token 的具体内容，由 JWT 标准中注册的声明、公共的声 明和私有的声明三部分组成。 
- Signature：Signature 是 Token 的签名部分，程序通过验证 Signature 是否合法，来 决定认证是否通过。

### 课后练习

- 思考下：使用 JWT 作为登录凭证，如何解决 token 注销问题？
  - JWT中可新增一个 valid 字段用于表示 token 是否有效，注销后则无效。
  - 删除客户端的缓存
- 思考下：Token 是存放在 LocalStorage 中好，还是存放在 Cookie 中好？
  - token还是存储在cookie中比较好，可由服务端保存，localstorage在纯前端中很容易泄露。
- 用UUID做凭证有很多问题： 
  - 1. 没有过期时间 
    2. UUID不包含用户信息 
    3. UUID可以伪造，Token有加密算法加密一些信息，更难伪造



## GO项目之IAM访问认证

应用认证常用的四种方式：Basic、Digest、OAuth、Bearer。

再来看下 IAM 项目是如何设计和实现认证功能的。 

IAM 项目用到了 Basic 认证和 Bearer 认证。其中，Basic 认证用在前端登陆的场景， Bearer 认证用在调用后端 API 服务的场景下。 

接下来，先来看下 IAM 项目认证功能的整体设计思路。 

### 如何设计 IAM 项目的认证功能？

在认证功能开发之前，要根据需求，认真考虑下如何设计认证功能，并在设计阶段通 过技术评审。那么先来看下，如何设计 IAM 项目的认证功能。 

首先，要梳理清楚认证功能的使用场景和需求。

- IAM 项目的 iam-apiserver 服务，提供了 IAM 系统的管理流功能接口，它的客户端可以是前端（这里也叫控制台），也可以是 App 端。 
- 为了方便用户在 Linux 系统下调用，IAM 项目还提供了 iamctl 命令行工具。 
- 为了支持在第三方代码中调用 iam-apiserver 提供的 API 接口，还支持了 API 调用。 
- 为了提高用户在代码中调用 API 接口的效率，IAM 项目提供了 Go SDK。

可以看到，iam-apiserver 有很多客户端，每种客户端适用的认证方式是有区别的。 

- 控制台、App 端需要登录系统，所以需要使用用户名：密码这种认证方式，也即 Basic 认证。
- iamctl、API 调用、Go SDK 因为可以不用登录系统，所以可以采用更安全的认证方 式：Bearer 认证。
- 同时，Basic 认证作为 iam-apiserver 已经集成的认证方式，仍然可以 供 iamctl、API 调用、Go SDK 使用。 

这里有个地方需要注意：如果 iam-apiserver 采用 Bearer Token 的认证方式，目前最受欢迎的 Token 格式是 JWT Token。而 JWT Token 需要密钥（后面统一用 secretKey 来指代），因此需要在 iam-apiserver 服务中为每个用户维护一个密钥，这样会增加开发和维护成本。 

业界有一个更好的实现方式：将 iam-apiserver 提供的 API 接口注册到 API 网关中，通过 API 网关中的 Token 认证功能，来实现对 iam-apiserver API 接口的认证。有很多 API 网关可供选择，例如腾讯云 API 网关、Tyk、Kong 等。 

这里需要注意：通过 iam-apiserver 创建的密钥对是提供给 iam-authz-server 使用 的。 另外，还需要调用 iam-authz-server 提供的 RESTful API 接口：/v1/authz，来进行资源授权。API 调用比较适合采用的认证方式是 Bearer 认证。

当然，/v1/authz也可以直接注册到 API 网关中。在实际的 Go 项目开发中，也是推荐 的一种方式。但在这里，为了展示实现 Bearer 认证的过程，iam-authz-server 自己实现 了 Bearer 认证。讲到 iam-authz-server Bearer 认证实现的时候，会详细介绍这一点。 

Basic 认证需要用户名和密码，Bearer 认证则需要密钥，所以 iam-apiserver 需要将用户名 / 密码、密钥等信息保存在后端的 MySQL 中，持久存储起来。 

在进行认证的时候，需要获取密码或密钥进行反加密，这就需要查询密码或密钥。查询密码或密钥有两种方式。

- 一种是在请求到达时查询数据库。因为数据库的查询操作延时高， 会导致 API 接口延时较高，所以不太适合用在数据流组件中。
- 另外一种是将密码或密钥缓存在内存中，这样请求到来时，就可以直接从内存中查询，从而提升查询速度，提高接口性能。 

但是，将密码或密钥缓存在内存中时，就要考虑内存和数据库的数据一致性，这会增加代码实现的复杂度。因为管控流组件对性能延时要求不那么敏感，而数据流组件则一定要实现非常高的接口性能，所以 iam-apiserver 在请求到来时查询数据库，而 iam-authz-server 则将密钥信息缓存在内存中。 

那在这里，可以总结出一张 IAM 项目的认证设计图：

![image-20211202225610976](IAM-document.assets/image-20211202225610976.png)

另外，为了将控制流和数据流区分开来，密钥的 CURD 操作也放在了 iam-apiserver 中， 但是 iam-authz-server 需要用到这些密钥信息。为了解决这个问题，目前的做法是：

- iam-authz-server 通过 gRPC API 请求 iam-apiserver，获取所有的密钥信息； 
- 当 iam-apiserver 有密钥更新时，会 Pub 一条消息到 Redis Channel 中。因为 iam-authz-server 订阅了同一个 Redis Channel，iam-authz-searver 监听到 channel 有新消息时，会获取、解析消息，并更新它缓存的密钥信息。这样，就能确保 iam-authz-server 内存中缓存的密钥和 iam-apiserver 中的密钥保持一致。

学到这里，可能会问：将所有密钥都缓存在 iam-authz-server 中，那岂不是要占用很大的内存？

- 别担心，这个问题也想过，并且计算好了：8G 的内存大概能保存约 8 千万个密钥信息，完全够用。
- 后期不够用的话，可以加大内存。 

不过这里还是有个小缺陷：如果 Redis down 掉，或者出现网络抖动，可能会造成 iam-apiserver 中和 iam-authz-server 内存中保存的密钥数据不一致，但这不妨碍学习认证功能的设计和实现。

至于如何保证缓存系统的数据一致性，会在新一期的特别放送里专门介绍下。 

最后注意一点：Basic 认证请求和 Bearer 认证请求都可能被截获并重放。所以，为了确保 Basic 认证和 Bearer 认证的安全性，和服务端通信时都需要配合使用 HTTPS 协议。

### IAM 项目是如何实现 Basic 认证的？ 

已经知道，IAM 项目中主要用了 Basic 和 Bearer 这两种认证方式。要支持 Basic 认证和 Bearer 认证，并根据需要选择不同的认证方式，这很容易想到使用设计模式中的策略模式来实现。

所以，在 IAM 项目中，将每一种认证方式都视作一个策略，通过选择不同的策略，来使用不同的认证方法。 IAM 项目实现了如下策略：

- auto 策略：该策略会根据 HTTP 头Authorization: Basic XX.YY.ZZ和 Authorization: Bearer XX.YY.ZZ自动选择使用 Basic 认证还是 Bearer 认证。 
- basic 策略：该策略实现了 Basic 认证。 
- jwt 策略：该策略实现了 Bearer 认证，JWT 是 Bearer 认证的具体实现。 
- cache 策略：该策略其实是一个 Bearer 认证的实现，Token 采用了 JWT 格式，因为 Token 中的密钥 ID 是从内存中获取的，所以叫 Cache 认证。这一点后面会详细介绍。

iam-apiserver 通过创建需要的认证策略，并加载到需要认证的 API 路由上，来实现 API 认证。具体代码如下：

```go
jwtStrategy, _ := newJWTAuth().(auth.JWTStrategy)
g.POST("/login", jwtStrategy.LoginHandler)
g.POST("/logout", jwtStrategy.LogoutHandler)
// Refresh time can be longer than token timeout
g.POST("/refresh", jwtStrategy.RefreshHandler)
```

上述代码中，通过 newJWTAuth函数创建了auth.JWTStrategy类型的变量，该 变量包含了一些认证相关函数。

- LoginHandler：实现了 Basic 认证，完成登陆认证。 
- RefreshHandler：重新刷新 Token 的过期时间。 
- LogoutHandler：用户注销时调用。登陆成功后，如果在 Cookie 中设置了认证相关的信息，执行 LogoutHandler 则会清空这些信息。

下面，来分别介绍下 LoginHandler、RefreshHandler 和 LogoutHandler。

#### LoginHandler

1. LoginHandler

这里，来看下 LoginHandler Gin 中间件，该函数定义位于 github.com/appleboy/gin-jwt 包的 auth_jwt.go文件中。

> https://github.com/appleboy/gin-jwt/blob/master/auth_jwt.go

```go
// LoginHandler can be used by clients to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// Reply will be of the form {"token": "TOKEN"}.
func (mw *GinJWTMiddleware) LoginHandler(c *gin.Context) {
	if mw.Authenticator == nil {
		mw.unauthorized(c, http.StatusInternalServerError, mw.HTTPStatusMessageFunc(ErrMissingAuthenticatorFunc, c))
		return
	}

	data, err := mw.Authenticator(c)

	if err != nil {
		mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, c))
		return
	}

	// Create the token
	token := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))
	claims := token.Claims.(jwt.MapClaims)

	if mw.PayloadFunc != nil {
		for key, value := range mw.PayloadFunc(data) {
			claims[key] = value
		}
	}

	expire := mw.TimeFunc().Add(mw.Timeout)
	claims["exp"] = expire.Unix()
	claims["orig_iat"] = mw.TimeFunc().Unix()
	tokenString, err := mw.signedString(token)

	if err != nil {
		mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(ErrFailedTokenCreation, c))
		return
	}

	// set cookie
	if mw.SendCookie {
		expireCookie := mw.TimeFunc().Add(mw.CookieMaxAge)
		maxage := int(expireCookie.Unix() - mw.TimeFunc().Unix())

		if mw.CookieSameSite != 0 {
			c.SetSameSite(mw.CookieSameSite)
		}

		c.SetCookie(
			mw.CookieName,
			tokenString,
			maxage,
			"/",
			mw.CookieDomain,
			mw.SecureCookie,
			mw.CookieHTTPOnly,
		)
	}

	mw.LoginResponse(c, http.StatusOK, tokenString, expire)
}
```

从 LoginHandler 函数的代码实现中，可以知道，LoginHandler 函数会执行 Authenticator 函数，来完成 Basic 认证。如果认证通过，则会签发 JWT Token，并执 行 PayloadFunc 函数设置 Token Payload。如果设置了 SendCookie=true ，还会 在 Cookie 中添加认证相关的信息，例如 Token、Token 的生命周期等，最后执行 LoginResponse 方法返回 Token 和 Token 的过期时间。

Authenticator、PayloadFunc、LoginResponse这三个函数，是在创建 JWT 认 证策略时指定的。下面来分别介绍下。 

##### Authenticator 函数

先来看下 Authenticator 函数。Authenticator 函数从 HTTP Authorization Header 中 获取用户名和密码，并校验密码是否合法。

```go
// internal/apiserver/auth.go
// new version is changed
func (auth *jwtAuth) Authenticator() func(c *gin.Context) (interface{}, error) {
   return func(c *gin.Context) (interface{}, error) {
      var login loginInfo
      var err error
      // support header and body both
      if c.Request.Header.Get("Authorization") != "" {
        login, err = parseWithHeader(c)
      } else {
        login, err = parseWithBody(c)
      }
      if err != nil {
        return "", jwt.ErrFailedAuthentication
      }
      // Get the user information by the login username.
      user, err := store.Client().Users().Get(c, login.Username, metav1.GetOptio
      if err != nil {
        log.Errorf("get user information failed: %s", err.Error())
        return "", jwt.ErrFailedAuthentication
      }
      // Compare the login password with the user password.
      if err := user.Compare(login.Password); err != nil {
        return "", jwt.ErrFailedAuthentication
      }
      return user, nil
   }
}
```

Authenticator函数需要获取用户名和密码。它首先会判断是否有Authorization请求 头，如果有，则调用parseWithHeader函数获取用户名和密码，否则调用 parseWithBody从 Body 中获取用户名和密码。如果都获取失败，则返回认证失败错 误。

所以，IAM 项目的 Basic 支持以下两种请求方式：

```sh
$ curl -XPOST -H"Authorization: Basic YWRtaW46QWRtaW5AMjAyMQ==" http://127.0.0.1:8080/login # 用户名:密码通过base64加码后，通过HTTP Authorization Header进行传递，因为密码非明文，建议使用这种方式。

$ curl -s -XPOST -H"Content-Type: application/json" -d'{"username":"admin","password":"Admin@2021"}' http://127.0.0.1:8080/login # 用户名和密码在HTTP Body中传递，因为密码是明文，所以这里不建议实际开发中，使用这种方式。
```

###### parseWithHeader 函数

这里，来看下 parseWithHeader 是如何获取用户名和密码的。假设请求为：

```sh
$ curl -XPOST -H"Authorization: Basic YWRtaW46QWRtaW5AMjAyMQ==" http://127.0.0.1:8080/login
```

其中，YWRtaW46QWRtaW5AMjAyMQ==值由以下命令生成：

```sh
$ echo -n 'admin:Admin@2021'|base64
YWRtaW46QWRtaW5AMjAyMQ==
```

parseWithHeader实际上执行的是上述命令的逆向步骤：

- 获取Authorization头的值，并调用 strings.SplitN 函数，获取一个切片变量 auth， 其值为 ["Basic","YWRtaW46QWRtaW5AMjAyMQ=="] 。 
- 将YWRtaW46QWRtaW5AMjAyMQ==进行 base64 解码，得到admin:Admin@2021。 
- 调用strings.SplitN函数获取 admin:Admin@2021 ，得到用户名为admin，密码为 Admin@2021。

###### parseWithBody 函数

parseWithBody 则是调用了 Gin 的ShouldBindJSON函数，来从 Body 中解析出用户名和密码。 

获取到用户名和密码之后，程序会从数据库中查询出该用户对应的加密后的密码，这里假设是xxxx。

最后authenticator函数调用user.Compare来判断 xxxx 是否和通过 user.Compare加密后的字符串相匹配，如果匹配则认证成功，否则返回认证失败。

##### PayloadFunc 函数

再来看下PayloadFunc函数：

```go
func (auth *jwtAuth) PayloadFunc() func(data interface{}) jwt.MapClaims {
   return func(data interface{}) jwt.MapClaims {
      claims := jwt.MapClaims{
         "iss": APIServerIssuer,
         "aud": APIServerAudience,
      }
      if u, ok := data.(*v1.User); ok {
         claims[jwt.IdentityKey] = u.Name
         claims["sub"] = u.Name
      }

      return claims
   }
}
```

PayloadFunc 函数会设置 JWT Token 中 Payload 部分的 iss、aud、sub、identity 字 段，供后面使用。 

##### LoginResponse 函数

再来看下刚才说的第三个函数，LoginResponse 函数：

```go
func (auth *jwtAuth) LoginResponse() func(c *gin.Context, code int, token string, expire time.Time) {
   return func(c *gin.Context, code int, token string, expire time.Time) {
      c.JSON(http.StatusOK, gin.H{
         "token":  token,
         "expire": expire.Format(time.RFC3339),
      })
   }
}
```

该函数用来在 Basic 认证成功之后，返回 Token 和 Token 的过期时间给调用者：

```sh
$ curl -XPOST -H"Authorization: Basic YWRtaW46QWRtaW5AMjAyMQ==" http://127.0.0.1:8080/login
{"expire":"2021-12-03T23:38:50+08:00","token":"XX.YY.ZZ"}
```

登陆成功后，iam-apiserver 会返回 Token 和 Token 的过期时间，前端可以将这些信息 缓存在 Cookie 中或 LocalStorage 中，之后的请求都可以使用 Token 来进行认证。

使用 Token 进行认证，不仅能够提高认证的安全性，还能够避免查询数据库，从而提高认证效 率。

#### RefreshHandler

2. RefreshHandler

RefreshHandler 函数会先执行 Bearer 认证，如果认证通过，则会重新签发 Token。

```go
// RefreshHandler can be used to refresh a token. The token still needs to be valid on refresh.
// Shall be put under an endpoint that is using the GinJWTMiddleware.
// Reply will be of the form {"token": "TOKEN"}.
func (mw *GinJWTMiddleware) RefreshHandler(c *gin.Context) {
 tokenString, expire, err := mw.RefreshToken(c)
 if err != nil {
  mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, c))
  return
 }

 mw.RefreshResponse(c, http.StatusOK, tokenString, expire)
}
```

#### LogoutHandler

3. LogoutHandler

最后，来看下LogoutHandler函数：

```go
// LogoutHandler can be used by clients to remove the jwt cookie (if set)
func (mw *GinJWTMiddleware) LogoutHandler(c *gin.Context) {
 // delete auth cookie
 if mw.SendCookie {
  if mw.CookieSameSite != 0 {
   c.SetSameSite(mw.CookieSameSite)
  }

  c.SetCookie(
   mw.CookieName,
   "",
   -1,
   "/",
   mw.CookieDomain,
   mw.SecureCookie,
   mw.CookieHTTPOnly,
  )
 }

 mw.LogoutResponse(c, http.StatusOK)
}
```

可以看到，LogoutHandler 其实是用来清空 Cookie 中 Bearer 认证相关信息的。 

#### 总结

最后，做个总结：Basic 认证通过用户名和密码来进行认证，通常用在登陆接口 /login 中。用户登陆成功后，会返回 JWT Token，前端会保存该 JWT Token 在浏览器的 Cookie 或 LocalStorage 中，供后续请求使用。

后续请求时，均会携带该 Token，以完成 Bearer 认证。另外，有了登陆接口，一般还会配套 /logout 接口和 /refresh 接口，分别用来进行注销和刷新 Token。 

这里可能会问，为什么要刷新 Token？因为通过登陆接口签发的 Token 有过期时间，有了刷新接口，前端就可以根据需要，自行刷新 Token 的过期时间。

过期时间可以通过 iam-apiserver 配置文件(./configs/iam-apiserver.yaml)的 jwt.timeout 配置项来指定。登陆后签发 Token 时，使用的密钥（secretKey）由 jwt.key配置项来指定。

### IAM 项目是如何实现 Bearer 认证的？ 

上面介绍了 Basic 认证。这里，再来介绍下 IAM 项目中 Bearer 认证的实现方式。 

IAM 项目中有两个地方实现了 Bearer 认证，分别是 iam-apiserver 和 iam-authz-server。下面来分别介绍下它们是如何实现 Bearer 认证的。 

#### iam-authz-server Bearer 认证实现 

先来看下 iam-authz-server 是如何实现 Bearer 认证的。 iam-authz-server 通过在 /v1 路由分组中加载 cache 认证中间件来使用 cache 认证策 略：

```go
// ./internal/authzserver/router.go
auth := newCacheAuth()
apiv1 := g.Group("/v1", auth.AuthFunc())
```

##### newCacheAuth 函数

来看下 newCacheAuth函数：

```go
func newCacheAuth() middleware.AuthStrategy {
   return auth.NewCacheStrategy(getSecretFunc())
}

func getSecretFunc() func(string) (auth.Secret, error) {
   return func(kid string) (auth.Secret, error) {
      cli, err := cache.GetCacheInsOr(nil)
      if err != nil || cli == nil {
         return auth.Secret{}, errors.Wrap(err, "get cache instance failed")
      }

      secret, err := cli.GetSecret(kid)
      if err != nil {
         return auth.Secret{}, err
      }

      return auth.Secret{
         Username: secret.Username,
         ID:       secret.SecretId,
         Key:      secret.SecretKey,
         Expires:  secret.Expires,
      }, nil
   }
}
```

newCacheAuth 函数调用auth.NewCacheStrategy创建了一个 cache 认证策略，创建 时传入了getSecretFunc函数，该函数会返回密钥的信息。密钥信息包含了以下字段：

```go
// Secret contains the basic information of the secret key.
type Secret struct {
   Username string
   ID       string
   Key      string
   Expires  int64
}
```

##### AuthFunc 函数

再来看下 cache 认证策略实现的 AuthFunc方法：

```go
// AuthFunc defines cache strategy as the gin authentication middleware.
func (cache CacheStrategy) AuthFunc() gin.HandlerFunc {
   return func(c *gin.Context) {
      header := c.Request.Header.Get("Authorization")
      if len(header) == 0 {
         core.WriteResponse(c, errors.WithCode(code.ErrMissingHeader, "Authorization header cannot be empty."), nil)
         c.Abort()

         return
      }

      var rawJWT string
      // Parse the header to get the token part.
      fmt.Sscanf(header, "Bearer %s", &rawJWT)

      // Use own validation logic, see below
      var secret Secret

      claims := &jwt.MapClaims{}
      // Verify the token
      parsedT, err := jwt.ParseWithClaims(rawJWT, claims, func(token *jwt.Token) (interface{}, error) {
         // Validate the alg is HMAC signature
         if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
         }

         kid, ok := token.Header["kid"].(string)
         if !ok {
            return nil, ErrMissingKID
         }

         var err error
         secret, err = cache.get(kid)
         if err != nil {
            return nil, ErrMissingSecret
         }

         return []byte(secret.Key), nil
      }, jwt.WithAudience(AuthzAudience))
      if err != nil || !parsedT.Valid {
         core.WriteResponse(c, errors.WithCode(code.ErrSignatureInvalid, err.Error()), nil)
         c.Abort()

         return
      }

      if KeyExpired(secret.Expires) {
         tm := time.Unix(secret.Expires, 0).Format("2006-01-02 15:04:05")
         core.WriteResponse(c, errors.WithCode(code.ErrExpired, "expired at: %s", tm), nil)
         c.Abort()

         return
      }

      c.Set(middleware.UsernameKey, secret.Username)
      c.Next()
   }
}

// KeyExpired checks if a key has expired, if the value of user.SessionState.Expires is 0, it will be ignored.
func KeyExpired(expires int64) bool {
   if expires >= 1 {
      return time.Now().After(time.Unix(expires, 0))
   }

   return false
}
```

AuthFunc 函数依次执行了以下四大步来完成 JWT 认证，每一步中又有一些小步骤，下面来一起看看。 

- 第一步，从 Authorization: Bearer XX.YY.ZZ 请求头中获取 XX.YY.ZZ，XX.YY.ZZ 即为 JWT Token。 

- 第二步，调用 github.com/dgrijalva/jwt-go 包提供的 ParseWithClaims 函数，该函数会依次执行下面四步操作。 

  - 调用 ParseUnverified 函数，依次执行以下操作： 

  - 从 Token 中获取第一段 XX，base64 解码后得到 JWT Token 的 `Header{“alg”:“HS256”,“kid”:“a45yPqUnQ8gljH43jAGQdRo0bXzNLjlU0hxa” ,“typ”:“JWT”}`。

  - 从 Token 中获取第一段 YY，base64 解码后得到 JWT Token 的 `Payload{“aud”:“iam.authz.marmotedu.com”,“exp”:1625104314,“iat”:16250 97114,“iss”:“iamctl”,“nbf”:1625097114}`。 

  - 根据 Token Header 中的 alg 字段，获取 Token 加密函数。 

  - 最终 ParseUnverified 函数会返回 Token 类型的变量，Token 类型包含 Method、 Header、Claims、Valid 这些重要字段，这些字段会用于后续的认证步骤中。 

  - 调用传入的 keyFunc 获取密钥，这里来看下 keyFunc 的实现：

    - ```go
      func(token *jwt.Token) (interface{}, error) {
         // Validate the alg is HMAC signature
         if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
         }
      
         kid, ok := token.Header["kid"].(string)
         if !ok {
            return nil, ErrMissingKID
         }
      
         var err error
         secret, err = cache.get(kid)
         if err != nil {
            return nil, ErrMissingSecret
         }
      
         return []byte(secret.Key), nil
      }
      ```

    - 可以看到，keyFunc 接受 *Token 类型的变量，并获取 Token Header 中的 kid，kid 即为密钥 ID：secretID。

    - 接着，调用 cache.get(kid) 获取密钥 secretKey。cache.get 函数 即为 getSecretFunc，getSecretFunc 函数会根据 kid，从内存中查找密钥信息，密钥信息中包含了 secretKey。

  - 从 Token 中获取 Signature 签名字符串 ZZZ，也即 Token 的第三段。

  - 获取到 secretKey 之后，token.Method.Verify(./dgrijalva/jwt-go/v4@v4.0.0-preview1/parser.go:38) 验证 Signature 签名字符串 ZZZ，也即 Token 的第三段是否合法。token.Method.Verify 实际上是使用了相同的加密算法和相 同的 secretKey 加密 XX.YY 字符串。假设加密之后的字符串为 WW，接下来会用 WW 和 ZZ base64 解码后的字符串进行比较，如果相等则认证通过，如果不相等则认证失 败。

- 第三步，调用 KeyExpired，验证 secret 是否过期。secret 信息中包含过期时间，只需要拿该过期时间和当前时间对比就行。 

- 第四步，设置 HTTP Header username: colin。 

到这里，iam-authz-server 的 Bearer 认证分析就完成了。 

##### 总结

来做个总结：iam-authz-server 通过加载 Gin 中间件的方式，在请求/v1/authz接 口时进行访问认证。因为 Bearer 认证具有过期时间，而且可以在认证字符串中携带更多有 用信息，还具有不可逆加密等优点，所以 /v1/authz 采用了 Bearer 认证，Token 格式采 用了 JWT 格式，这也是业界在 API 认证中最受欢迎的认证方式。 

Bearer 认证需要 secretID 和 secretKey，这些信息会通过 gRPC 接口调用，从 iam-apisaerver 中获取，并缓存在 iam-authz-server 的内存中供认证时查询使用。

当请求来临时，iam-authz-server Bearer 认证中间件从 JWT Token 中解析出 Header， 并从 Header 的 kid 字段中获取到 secretID，根据 secretID 查找到 secretKey，最后使用 secretKey 加密 JWT Token 的 Header 和 Payload，并与 Signature 部分进行对比。如果相等，则认证通过；如果不等，则认证失败。 

#### iam-apiserver Bearer 认证实现 

再来看下 iam-apiserver 的 Bearer 认证。 iam-apiserver 的 Bearer 认证通过以下代码（位于 router.go文件中）指定使用了 auto 认证策略：

```go
v1.Use(auto.AuthFunc())
```

来看下 auto.AuthFunc()的实现：

```go
// AuthFunc defines auto strategy as the gin authentication middleware.
func (a AutoStrategy) AuthFunc() gin.HandlerFunc {
   return func(c *gin.Context) {
      operator := middleware.AuthOperator{}
      authHeader := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

      if len(authHeader) != authHeaderCount {
         core.WriteResponse(
            c,
            errors.WithCode(code.ErrInvalidAuthHeader, "Authorization header format is wrong."),
            nil,
         )
         c.Abort()

         return
      }

      switch authHeader[0] {
      case "Basic":
         operator.SetStrategy(a.basic)
      case "Bearer":
         operator.SetStrategy(a.jwt)
         // a.JWT.MiddlewareFunc()(c)
      default:
         core.WriteResponse(c, errors.WithCode(code.ErrSignatureInvalid, "unrecognized Authorization header."), nil)
         c.Abort()

         return
      }

      operator.AuthFunc()(c)

      c.Next()
   }
}
```

从上面代码中可以看到，AuthFunc 函数会从 Authorization Header 中解析出认证方式是 Basic 还是 Bearer。如果是 Bearer，就会使用 JWT 认证策略；如果是 Basic，就会使用 Basic 认证策略。 

#####  JWT 认证策略 AuthFunc 函数

再来看下 JWT 认证策略的 AuthFunc函数实现：

```go
// AuthFunc defines jwt bearer strategy as the gin authentication middleware.
func (j JWTStrategy) AuthFunc() gin.HandlerFunc {
   return j.MiddlewareFunc()
}
```

跟随代码，可以定位到MiddlewareFunc函数最终调用了 github.com/appleboy/gin-jwt包 GinJWTMiddleware 结构体的 middlewareImpl 方法：

```go
func (mw *GinJWTMiddleware) middlewareImpl(c *gin.Context) {
 claims, err := mw.GetClaimsFromJWT(c)
 if err != nil {
  mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, c))
  return
 }

 if claims["exp"] == nil {
  mw.unauthorized(c, http.StatusBadRequest, mw.HTTPStatusMessageFunc(ErrMissingExpField, c))
  return
 }

 if _, ok := claims["exp"].(float64); !ok {
  mw.unauthorized(c, http.StatusBadRequest, mw.HTTPStatusMessageFunc(ErrWrongFormatOfExp, c))
  return
 }

 if int64(claims["exp"].(float64)) < mw.TimeFunc().Unix() {
  mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(ErrExpiredToken, c))
  return
 }

 c.Set("JWT_PAYLOAD", claims)
 identity := mw.IdentityHandler(c)

 if identity != nil {
  c.Set(mw.IdentityKey, identity)
 }

 if !mw.Authorizator(identity, c) {
  mw.unauthorized(c, http.StatusForbidden, mw.HTTPStatusMessageFunc(ErrForbidden, c))
  return
 }

 c.Next()
}
```

分析上面的代码，可以知道，middlewareImpl 的 Bearer 认证流程为：

- 第一步：调用GetClaimsFromJWT函数，从 HTTP 请求中获取 Authorization Header， 并解析出 Token 字符串，进行认证，最后返回 Token Payload。 
- 第二步：校验 Payload 中的exp是否超过当前时间，如果超过就说明 Token 过期，校验不通过。 
- 第三步：给 gin.Context 中添加 JWT_PAYLOAD 键，供后续程序使用（当然也可能用不到）。 
- 第四步：通过以下代码，在 gin.Context 中添加 IdentityKey 键，IdentityKey 键可以在创建 GinJWTMiddleware 结构体时指定，这里设置为middleware.UsernameKey，也就是 username。

```go
identity := mw.IdentityHandler(c)

if identity != nil {
 c.Set(mw.IdentityKey, identity)
}
```

IdentityKey 键的值由 IdentityHandler 函数返回，IdentityHandler 函数为：

```go
func(c *gin.Context) interface{} {
   claims := jwt.ExtractClaims(c)

   return claims[jwt.IdentityKey]
}
```

上述函数会从 Token 的 Payload 中获取 identity 域的值，identity 域的值是在签发 Token 时指定的，它的值其实是用户名，可以查看 payloadFunc函数了解。 

- 第五步：接下来，会调用Authorizator方法，Authorizator是一个 callback 函数，成 功时必须返回真，失败时必须返回假。Authorizator也是在创建 GinJWTMiddleware 时指定的，例如：

```go
func authorizator() func(data interface{}, c *gin.Context) bool {
   return func(data interface{}, c *gin.Context) bool {
      if v, ok := data.(string); ok {
         // c.Set(log.KeyUsername, v)
         // c.Set(log.KeyRequestID, v)
         log.L(c).Infof("user `%s` is authenticated.", v)

         return true
      }

      return false
   }
}
```

authorizator 函数返回了一个匿名函数，匿名函数在认证成功后，会打印一条认证成功 日志。 

### IAM 项目认证功能设计技巧 

在设计 IAM 项目的认证功能时，也运用了一些技巧。

#### 技巧 1：面向接口编程 

在使用 NewAutoStrategy函数创建 auto 认证策略时，传入了 BasicStrategy、 JWTStrategy 接口类型的参数，这意味着 Basic 认证和 Bearer 认证都可以有不同的实 现，这样后期可以根据需要扩展新的认证方式。 

#### 技巧 2：使用抽象工厂模式 

auth.go文件中，通过 newBasicAuth、newJWTAuth、newAutoAuth 创建认证策略 时，返回的都是接口。通过返回接口，可以在不公开内部实现的情况下，让调用者使用提供的各种认证功能。 

#### 技巧 3：使用策略模式 

在 auto 认证策略中，会根据 HTTP 请求头Authorization: XXX X.Y.X中的 XXX 来选择并设置认证策略（Basic 或 Bearer）。具体可以查看AutoStrategy的 AuthFunc函数：

```go
// AuthFunc defines auto strategy as the gin authentication middleware.
func (a AutoStrategy) AuthFunc() gin.HandlerFunc {
   return func(c *gin.Context) {
      operator := middleware.AuthOperator{}
      authHeader := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

      ...

      switch authHeader[0] {
      case "Basic":
         operator.SetStrategy(a.basic)
      case "Bearer":
         operator.SetStrategy(a.jwt)
         // a.JWT.MiddlewareFunc()(c)
      default:
         core.WriteResponse(c, errors.WithCode(code.ErrSignatureInvalid, "unrecognized Authorization header."), nil)
         c.Abort()

         return
      }

      operator.AuthFunc()(c)

      c.Next()
   }
}
```

上述代码中，如果是 Basic，则设置为 Basic 认证方法 `operator.SetStrategy(a.basic)`；如果是 Bearer，则设置为 Bearer 认证方法 `operator.SetStrategy(a.jwt)`。 

SetStrategy方法的入参是 AuthStrategy 类型 的接口，都实现了`AuthFunc() gin.HandlerFunc `函数，用来进行认证，所以最后调用operator.AuthFunc()(c)即可完成认证。 

### 总结 

在 IAM 项目中，iam-apiserver 实现了 Basic 认证和 Bearer 认证，iam-authz-server 实现了 Bearer 认证。重点介绍了 iam-apiserver 的认证实现。 

用户要访问 iam-apiserver，首先需要通过 Basic 认证，认证通过之后，会返回 JWT Token 和 JWT Token 的过期时间。前端将 Token 缓存在 LocalStorage 或 Cookie 中， 后续的请求都通过 Token 来认证。 

执行 Basic 认证时，iam-apiserver 会从 HTTP Authorization Header 中解析出用户名和密码，将密码再加密，并和数据库中保存的值进行对比。如果不匹配，则认证失败，否则认证成功。认证成功之后，会返回 Token，并在 Token 的 Payload 部分设置用户名， Key 为 username 。 

执行 Bearer 认证时，iam-apiserver 会从 JWT Token 中解析出 Header 和 Payload，并从 Header 中获取加密算法。接着，用获取到的加密算法和从配置文件中获取到的密钥对 Header.Payload 进行再加密，得到 Signature，并对比两次的 Signature 是否相等。如果不相等，则返回 HTTP 401 Unauthorized 错误；如果相等，接下来会判断 Token 是否过期，如果过期则返回认证不通过，否则认证通过。认证通过之后，会将 Payload 中的 username 添加到 gin.Context 类型的变量中，供后面的业务逻辑使用。 

绘制了整个流程的示意图，可以对照着再回顾一遍。

![image-20211203013113198](IAM-document.assets/image-20211203013113198.png)

### 课后练习

- 走读github.com/appleboy/gin-jwt包的 GinJWTMiddleware 结构体的 GetClaimsFromJWT 方法，分析一下：GetClaimsFromJWT 方法是如何从 gin.Context 中解析出 Token，并进行认证的？ 
- 思考下，iam-apiserver 和 iam-authzserver 是否可以使用同一个认证策略？如果可以，又该如何实现？
  - am-apiserver和iam-authz-server的api的认证功能 其实都应该放到网关来实现的，本文之所以由iam项目亲自来实现就是为了方便讲解认证 的具体实现方法。



## GO项目之权限模型

在开始讲解如何开发服务之前，先来介绍一个比较重要的背景知 识：权限模型。 

在研发生涯中，应该会遇到这样一种恐怖的操作：张三因为误操作删除了李四的资源。在刷新闻时，也可能会刷到这么一个爆款新闻：某某程序员删库跑路。操作之所以恐怖，新闻之所以爆款，是因为这些行为往往会带来很大的损失。 

那么如何避免这些风险呢？答案就是对资源做好权限管控，这也是项目开发中绕不开的话题。腾讯云会强制要求所有的云产品都对接访问管理（CAM） 服务（阿里云也有这种要求），之所以这么做，是因为保证资源的安全是一件非常非常重要的事情。

可以说，保证应用的资源安全，已经成为一个应用的必备能力。作为开发人员，也一定 要知道如何保障应用的资源安全。那么如何才能保障资源的安全呢？至少需要掌 握下面这两点：

- 权限模型：需要了解业界成熟的权限模型，以及这些模型的适用场景。只有具备足够 宽广的知识面和视野，才能避免闭门造车，设计出优秀的资源授权方案。 
- 编码实现：选择或设计出了优秀的资源授权方案后，就要编写代码实现该方案。这门 课的 IAM 应用，就是一个资源授权方案的落地项目。可以通过对 IAM 应用的学习， 来掌握如何实现一个资源授权系统。

无论是第一点还是第二点，都需要掌握基本的权限模型知识。那么这一讲，就来介绍 下业界优秀的权限模型，以及这些模型的适用场景，以使今后设计出更好的资源授权系 统。 

### 权限相关术语介绍 

在介绍业界常见的权限模型前，先来看下在权限模型中出现的术语。把常见的术语 总结在了下面的表格里：

![image-20211205111936478](IAM-document.assets/image-20211205111936478.png)

为了方便理解，这一讲分别用用户、操作和资源来替代 Subject、Action 和 Object。 

### 权限模型介绍 

接下来，就详细介绍下一些常见的权限模型，让今后在设计权限系统时，能够根据需 求选择合适的权限模型。 不同的权限模型具有不同的特点，可以满足不同的需求。常见的权限模型有下面这 5 种：

- 权限控制列表（ACL，Access Control List）。 
- 自主访问控制（DAC，Discretionary Access Control）。 
- 强制访问控制（MAC，Mandatory Access Control）。 
- 基于角色的访问控制（RBAC，Role-Based Access Control）。 
- 基于属性的权限验证（ABAC，Attribute-Based Access Control）。

这里先简单介绍下这 5 种权限模型。

- ACL 是一种简单的权限模型；
- DAC 基于 ACL，将权限下放给具有此权限的主题；但 DAC 因为权限下放，导致它对权限的控制过于分散，
- 为了弥补 DAC 的这个缺陷，诞生了 MAC 权限模型。 DAC 和 MAC 都是基于 ACL 的权限模型。
- ACL 及其衍生的权限模型可以算是旧时代的权 限模型，灵活性和功能性都满足不了现代应用的权限需求，所以诞生了 RBAC。RBAC 也 是迄今为止最为普及的权限模型。 但是，随着组织和应用规模的增长，所需的角色数量越来越多，变得难以管理，进而导致角色爆炸和职责分离（SoD）失败。
- 最后，引入了一种新的、更动态的访问控制形式，称为基于属性的访问控制，也就是 ABAC。ABAC 被一些人看作是权限系统设计的未来。腾讯云的 CAM、AWS 的 IAM、阿里云的 RAM 都是 ABAC 类型的权限访问服务。 

接下来，详细介绍这些权限模型的基本概念。 

#### 简单的权限模型：权限控制列表（ACL） 

ACL（Access Control List，权限控制列表），用来判断用户是否可以对资源做特定的操 作。例如，允许 Colin 创建文章的 ACL 策略为：

```yaml
Subject: Colin
Action: Create
Object: Article
```

在 ACL 权限模型下，权限管理是围绕资源 Object 来设定的，ACL 权限模型也是比较简单 的一种模型。 

#### 基于 ACL 下放权限的权限模型：自主访问控制（DAC） 

DAC (Discretionary Access Control，自主访问控制)，是 ACL 的扩展模型，灵活性更强。

使用这种模型，不仅可以判断 Subject 是否可以对 Object 做 Action 操作，同时也能让 Subject 将 Object、Action 的相同权限授权给其他的 Subject。例如，Colin 可以创建 文章：

```yaml
Subject: Colin
Action: Create
Object: Article
```

因为 Colin 具有创建文章的权限，所以 Colin 也可以授予 James 创建文章的权限：

```yaml
Subject: James
Action: Create
Object: Article
```

经典的 ACL 模型权限集中在同一个 Subject 上，缺乏灵活性，为了加强灵活性，在 ACL 的基础上，DAC 模型将权限下放，允许拥有权限的 Subject 自主地将权限授予其他 Subject。 

#### 基于 ACL 且安全性更高的权限模型：强制访问控制（MAC） 

MAC (Mandatory Access Control，强制访问控制)，是 ACL 的扩展模型，安全性更高。 

MAC 权限模型下，Subject 和 Object 同时具有安全属性。在做授权时，需要同时满足两 点才能授权通过：

- Subject 可以对 Object 做 Action 操作。 
- Object 可以被 Subject 做 Action 操作。

例如，设定了“Colin 和 James 可以创建文章”这个 MAC 策略：

```yaml
Subject: Colin
Action: Create
Object: Article

Subject: James
Action: Create
Object: Article
```

还有另外一个 MAC 策略“文章可以被 Colin 创建”：

```yaml
Subject: Article
Action: Create
Object: Colin
```

在上述策略中，Colin 可以创建文章，但是 James 不能创建文章，因为第二条要求没有满 足。 

这里需要注意，在 ACL 及其扩展模型中，Subject 可以是用户，也可以是组或群组。

 ACL、DAC 和 MAC 是旧时代的权限控制模型，无法满足现代应用对权限控制的需求，于是诞生了新时代的权限模型：RBAC 和 ABAC。 

#### 最普及的权限模型：基于角色的访问控制（RBAC） 

RBAC (Role-Based Access Control，基于角色的访问控制)，引入了 Role（角色）的概 念，并且将权限与角色进行关联。用户通过扮演某种角色，具有该角色的所有权限。具体如下图所示：

![image-20211205113333335](IAM-document.assets/image-20211205113333335.png)

如图所示，每个用户关联一个或多个角色，每个角色关联一个或多个权限，每个权限又包 含了一个或者多个操作，操作包含了对资源的操作集合。通过用户和权限解耦，可以实现非常灵活的权限管理。例如，可以满足以下两个权限场景： 

- 第一，可以通过角色批量给一个用户授权。
  - 例如，公司新来了一位同事，需要授权虚拟机 的生产、销毁、重启和登录权限。这时候，可以将这些权限抽象成一个运维角色。
  - 如果再有新同事来，就可以通过授权运维角色，直接批量授权这些权限，不用一个个地给用 户授权这些权限。 
- 第二，可以批量修改用户的权限。
  - 例如，有很多用户，同属于运维角色，这时候对运 维角色的任何权限变更，就相当于对运维角色关联的所有用户的权限变更，不用一个个去修改这些用户的权限。 

RBAC 又分为 RBAC0、RBAC1、RBAC2、RBAC3。

- RBAC0 是 RBAC 的核心思想， 
- RBAC1 是基于 RBAC 的角色分层模型，
- RBAC2 增加了 RBAC 的约束模型。
- 而 RBAC3， 其实相当于 RBAC1 + RBAC2。 

下面来详细介绍下这四种 RBAC。

##### RBAC0

RBAC0：基础模型，只包含核心的四要素，也就是用户（User）、角色（Role）、权限 （Permission：Objects-Operations）、会话（Session）。

用户和角色可以是多对多的 关系，权限和角色也是多对多的关系。 

##### RBAC1

RBAC1：包括了 RBAC0，并且添加了角色继承。

角色继承，即角色可以继承自其他角色， 在拥有其他角色权限的同时，还可以关联额外的权限。 

##### RBAC2

RBAC2：包括 RBAC0，并且添加了约束。具有以下核心特性：

- 互斥约束：包括互斥用户、互斥角色、互斥权限。
  - 同一个用户不能拥有相互排斥的角色，两个互斥角色不能分配一样的权限集，互斥的权限不能分配给同一个角色，在 Session 中，同一个角色不能拥有互斥权限。 
- 基数约束：一个角色被分配的用户数量受限，它指的是有多少用户能拥有这个角色。
  - 例如，一个角色是专门为公司 CEO 创建的，那这个角色的数量就是有限的。 
- 先决条件角色：指要想获得较高的权限，要首先拥有低一级的权限。
  - 例如，先有副总经理权限，才能有总经理权限。 
- 静态职责分离(Static Separation of Duty)：用户无法同时被赋予有冲突的角色。 
- 动态职责分离(Dynamic Separation of Duty)：用户会话中，无法同时激活有冲突的角 色。

##### RBAC3

RBAC3：全功能的 RBAC，合并了 RBAC0、RBAC1、RBAC2。

##### 模拟 DAC 和 MAC

此外，RBAC 也可以很方便地模拟出 DAC 和 MAC 的效果。例如，有 write article 和 manage article 的权限：

```yaml
Permission:
  - Name: write_article
    - Effect: "allow"
    - Action: ["Create", "Update", "Read"]
    - Object: ["Article"]
  - Name: manage_article
    - Effect: "allow"
    - Action: ["Delete", "Read"]
    - Object: ["Article"]
```

同时，也有一个 Writer 和 Manager 的角色，Writer 具有 write article 权限， Manager 具有 manage article 权限：

```yaml
Role:
  - Name: Writer
    Permissions:
    	- write_article
  - Name: Manager
    Permissions:
    	- manage_article
  - Name: CEO
    Permissions:
      - write_article
      - manage_article
```

接下来，对 Colin 用户授予 Writer 角色：

```yaml
Subject: Colin
Roles:
	- Writer
```

那么现在 Colin 就具有 Writer 角色的所有权限 write_article，write_article 权限可以创建 文章。

```yaml
Subject: James
Roles:
  - Writer
  - Manager
```

那么现在 James 就具有 Writer 角色和  Manager 角色的所有权限 write_article，manage_article。write_article 权限可以创建文章，manage_article 权限可以管理文章。

#### 最强大的权限模型：基于属性的权限验证（ABAC） 

ABAC (Attribute-Based Access Control，基于属性的权限验证），规定了哪些属性的用 户可以对哪些属性的资源在哪些限制条件下进行哪些操作。

跟 RBAC 相比，ABAC 对权限的控制粒度更细，主要规定了下面这四类属性：

- 用户属性，例如性别、年龄、工作等。 
- 资源属性，例如创建时间、所属位置等。 
- 操作属性，例如创建、修改等。 
- 环境属性，例如来源 IP、当前时间等。

下面是一个 ABAC 策略：

```yaml
Subject:
  Name: Colin
  Department: Product
  Role: Writer
Action:
  - create
  - update
Resource:
  Type: Article
  Tag:
    - technology
    - software
  Mode:
    - draft
Contextual:
  IP: 10.0.0.10
```

上面权限策略描述的意思是，产品部门的 Colin 作为一个 Writer 角色，可以通过来源 IP 是 10.0.0.10 的客户端，创建和更新带有 technology 和 software 标签的草稿文章。 

这里提示一点：ABAC 有时也被称为 PBAC（Policy-Based Access Control）或 CBAC（Claims-Based Access Control）。 

这里，通过现实中的 ABAC 授权策略，理解 ABAC 权限模型。下面是一个腾讯云的 CAM 策略，也是一种 ABAC 授权模式：

```json
{
   "version":"2.0",
   "statement":[
      {
         "effect":"allow",
         "action":[
            "cos:List*",
            "cos:Get*",
            "cos:Head*",
            "cos:OptionsObject"
         ],
         "resource":"qcs::cos:ap-shanghai:uid/1250000000:Bucket1-1250000000/dir1/*",
         "condition":{
            "ip_equal":{
               "qcs:ip":[
                  "10.217.182.3/24",
                  "111.21.33.72/24"
               ]
            }
         }
      }
   ]
}
```

上面的授权策略表示：用户必须在 10.217.182.3/24 或者 111.21.33.72/24 网段才能调用 云 API（`cos:List*、cos:Get*、cos:Head*、cos:OptionsObject`），对 1250000000 用 户下的 dir1 目录下的文件进行读取操作。 

这里，ABAC 规定的四类属性分别是：

- 用户属性：用户为 1250000000。
- 资源属性：dir1 目录下的文件。 
- 操作属性：读取（`cos:List*、cos:Get*、cos:Head*、cos:OptionsObject` ）都是读取 API。 
- 环境属性：10.217.182.3/24 或者 111.21.33.72/24 网段。

### 相关开源项目 

上面介绍了权限模型的相关知识，但是现在如果真正去实现一个权限系统，可能 还是不知从何入手。 

在这里，列出了一些 GitHub 上比较优秀的开源项目，可以学习这些项目是如何落地 一个权限模型的，也可以基于这些项目进行二次开发，开发一个满足业务需求的权限系统。 

#### Casbin 

Casbin 是一个用 Go 语言编写的访问控制框架，功能强大，支持 ACL、RBAC、ABAC 等访问模型，很多优秀的权限管理系统都是基于 Casbin 来构建的。

Casbin 的核心功能都 是围绕着访问控制来构建的，不负责身份认证。如果以后老板让你实现一个权限管理系 统，Casbin 是一定要好好研究的开源项目。 

#### keto 

keto 是一个云原生权限控制服务，通过提供 REST API 进行授权，支持 RBAC、 ABAC、ACL、AWS IAM 策略、Kubernetes Roles 等权限模型，可以解决下面这些问 题：

- 是否允许某些用户修改此博客文章？ 
- 是否允许某个服务打印该文档？ 
- 是否允许 ACME 组织的成员修改其租户中的数据？ 
- 是否允许在星期一的下午 4 点到下午 5 点，从 IP 10.0.0.2 发出的请求执行某个 Job？

#### go-admin

go-admin 是一个基于 Gin + Vue + Element UI 的前后端分离权限管理系统脚手架， 它的访问控制模型采用了 Casbin 的 RBAC 访问控制模型，功能强大，包含了如下功能：

- 基础用户管理功能； 
- JWT 鉴权； 
- 代码生成器； 
- RBAC 权限控制； 
- 表单构建； 
- ……

该项目还支持 RESTful API 设计规范、Swagger 文档、GORM 库等。

go-admin 不仅是一个优秀的权限管理系统，也是一个优秀的、功能齐全的 Go 开源项目。在做项目开发 时，也可以参考该项目的构建思路。go-admin 管理系统自带前端，如下图所示。

![image-20211205120106831](IAM-document.assets/image-20211205120106831.png)

#### LyricTian/gin-admin 

gin-admin 类似于 go-admin，是一个基于 Gin+Gorm+Casbin+Wire 实现的权限管理脚手架，并自带前端，在做权限管理系统调研时，也非常值得参考。 

gin-admin 大量采用了 Go 后端开发常用的技术，比如 Gin、GORM、JWT 认证、 RESTful API、Logrus 日志包、Swagger 文档等。因此，在做 Go 后端服务开发时，也 可以学习该项目的构建方法。 

#### gin-vue-admin

gin-vue-admin 是一个基于 Gin 和 Vue 开发的全栈前后端分离的后台管理系统，集成了 JWT 鉴权、动态路由、动态菜单、Casbin 鉴权、表单生成器、代码生成器等功能。 

gin-vue-admin 集成了 RBAC 权限管理模型，界面如下图所示：

![image-20211205120320944](IAM-document.assets/image-20211205120320944.png)

#### 选择建议 

介绍了那么多优秀的开源项目，最后给一些选择建议。

如果想研究 ACL、 RBAC、ABAC 等权限模型如何落地，强烈建议学习 Casbin 项目，Casbin 目前有 近万的 GitHub star 数，处于活跃的开发状态。有很多项目在使用 Casbin，例如 go-admin、 gin-admin 、 gin-vue-admin 等。 

keto 类似于 Casbin，主要通过 Go 包的方式，对外提供授权能力。keto 也是一个非常优秀的权限类项目，当研究完 Casbin 后，如果还想再研究下其他授权类项目，建议读 下 keto 的源码。 

go-admin、gin-vue-admin、gin-admin 这 3 个都是基于 Casbin 的 Go 项目。其中， gin-vue-admin 是后台管理系统框架，里面包含了 RBAC 权限管理模块；go-admin 和 gin-admin 都是 RBAC 权限管理脚手架。所以，如果想找一个比较完整的 RBAC 授权系统（自带前后端），建议优先研究下 go-admin，如果还有精力，可以再研究下 gin-admin、gin-vue-admin。

### 总结 

这一讲，介绍了 5 种常见的权限模型。

- 其中，ACL 最简单，ABAC 最复杂，但是功能最强大，也最灵活。、
- RBAC 则介于二者之间。
- 对于一些云计算厂商来说，因为它们面临的授 权场景复杂多样，需要一个非常强大的授权模型，所以腾讯云、阿里云和 AWS 等云厂商 普遍采用了 ABAC 模型。

如果资源授权需求不复杂，可以考虑 RBAC；如果需要一个能满足复杂场景的资源授权系统，建议选择 ABAC，ABAC 的设计思路可以参考下腾讯云的 CAM、阿里云的 RAM 和 AWS 的 IAM。 

另外，如果想深入了解权限模型如何具体落地，建议阅读 Casbin 源码。 

### 课后练习

- 思考一下，如果公司需要实现一个授权中台系统，应该选用哪种权限模型来构建，来满足不同业务的不同需求？
  - ABAC
- 思考一下，如何将授权流程集成进统一接入层，例如 API 网关？
  - 写一个网关插件，当访问认证通过后，自动调用类似本iam项目的后端应用作资源鉴权。

# GO项目之IAM：iam-apiserver设计

接下来，就讲解下 IAM 应用的源码。 

在讲解过程中，不会去讲解具体如何 Code，但会讲解一些构建过程中的重点、难点，以 及 Code 背后的设计思路、想法。

IAM 项目有很多组件，这一讲，先来介绍下 IAM 项目的门面服务：iam-apiserver（管 理流服务）。先介绍下 iam-apiserver 的功能和使用方法，再介绍下 iam-apiserver 的代码实现。

### iam-apiserver 服务介绍 

iam-apiserver 是一个 Web 服务，通过一个名为 iam-apiserver 的进程，对外提供 RESTful API 接口，完成用户、密钥、策略三种 REST 资源的增删改查。

接下来，从功能和使用方法两个方面来具体介绍下。 

#### iam-apiserver 功能介绍 

可以通过 iam-apiserver 提供的 RESTful API 接口，来看下 iam-apiserver 具体提供的功能。iam-apiserver 提供的 RESTful API 接口可以分为四类，具体如下： 

##### 认证相关接口

![image-20211206234356553](IAM-document.assets/image-20211206234356553.png)

##### 用户相关接口

![image-20211206234423037](IAM-document.assets/image-20211206234423037.png)

##### 密钥相关接口

![image-20211206234510692](IAM-document.assets/image-20211206234510692.png)

##### 策略相关接口

![image-20211206234534329](IAM-document.assets/image-20211206234534329.png)

#### iam-apiserver 使用方法介绍 

上面介绍了 iam-apiserver 的功能，接下来就介绍下如何使用这些功能。 

可以通过不同的客户端来访问 iam-apiserver，例如前端、API 调用、SDK、iamctl 等。这些客户端最终都会执行 HTTP 请求，调用 iam-apiserver 提供的 RESTful API 接口。

所以，首先需要有一个顺手的 REST API 客户端工具来执行 HTTP 请求，完成开发测试。 

因为不同的开发者执行 HTTP 请求的方式、习惯不同，为了方便讲解，这里统一通过 cURL 工具来执行 HTTP 请求。接下来先介绍下 cURL 工具。 

##### cURL 工具

标准的 Linux 发行版都安装了 cURL 工具。cURL 可以很方便地完成 RESTful API 的调用场景，比如设置 Header、指定 HTTP 请求方法、指定 HTTP 消息体、指定权限认证信息 等。通过-v选项，也能输出 REST 请求的所有返回信息。cURL 功能很强大，有很多参 数，这里列出 cURL 工具常用的参数：

```sh
-X/--request [GET|POST|PUT|DELETE|…] 指定请求的 HTTP 方法
-H/--header 指定请求的 HTTP Header
-d/--data 指定请求的 HTTP 消息体（Body）
-v/--verbose 输出详细的返回信息
-u/--user 指定账号、密码
-b/--cookie 读取 cookie
```

此外，如果想使用带 UI 界面的工具，这里我推荐你使用 Insomnia 。 

##### Insomnia 工具

Insomnia 是一个跨平台的 REST API 客户端，与 Postman、Apifox 是一类工具，用于接 口管理、测试。Insomnia 功能强大，支持以下功能：

- 发送 HTTP 请求； 
- 创建工作区或文件夹； 
- 导入和导出数据； 
- 导出 cURL 格式的 HTTP 请求命令； 
- 支持编写 swagger 文档； 
- 快速切换请求； 
- URL 编码和解码。 
- …

Insomnia 界面如下图所示：

![image-20211206235413146](IAM-document.assets/image-20211206235413146.png)

当然了，也有很多其他优秀的带 UI 界面的 REST API 客户端，例如 Postman、Apifox 等，可以根据需要自行选择。 

##### secret 资源的 CURD

接下来，用对 secret 资源的 CURD 操作，来演示下如何使用 iam-apiserver 的功 能。需要执行 6 步操作。

1. 登录 iam-apiserver，获取 token。 
2. 创建一个名为 secret0 的 secret。 
3. 获取 secret0 的详细信息。 
4. 更新 secret0 的描述。
5. 获取 secret 列表。
6. 删除 secret0。

具体操作如下：

1. 登录 iam-apiserver，获取 token：

```sh
$ curl -s -XPOST -H"Authorization: Basic `echo -n 'admin:Admin@2021'|base64`" http://127.0.0.1:8080/login | jq -r .token

eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2Mzg4OTM0MDksImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2Mzg4MDcwMDksInN1YiI6ImFkbWluIn0.2vKRDOUDyp9Jyj_Lk73gZR54TPYK52SwbToC_tC8HNo
```

这里，为了便于使用，将 token 设置为环境变量：

```sh
TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2Mzg4OTM0MDksImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2Mzg4MDcwMDksInN1YiI6ImFkbWluIn0.2vKRDOUDyp9Jyj_Lk73gZR54TPYK52SwbToC_tC8HNo
```

2. 创建一个名为 secret0 的 secret：

```SH
$ curl -v -XPOST -H "Content-Type: application/json" -H"Authorization: Bearer ${TOKEN}" -d'{"metadata":{"name":"secret0"},"expires":0,"description":"admin secret"}' http://iam.api.marmotedu.com:8080/v1/secrets

Note: Unnecessary use of -X or --request, POST is already inferred.
*   Trying 127.0.0.1...
* TCP_NODELAY set
* Connected to iam.api.marmotedu.com (127.0.0.1) port 8080 (#0)
> POST /v1/secrets HTTP/1.1
> Host: iam.api.marmotedu.com:8080
> User-Agent: curl/7.61.1
> Accept: */*
> Content-Type: application/json
> Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJpYW0uYXBpLm1hcm1vdGVkdS5jb20iLCJleHAiOjE2Mzg4OTI3MTQsImlkZW50aXR5IjoiYWRtaW4iLCJpc3MiOiJpYW0tYXBpc2VydmVyIiwib3JpZ19pYXQiOjE2Mzg4MDYzMTQsInN1YiI6ImFkbWluIn0.4vSotQOtE8SUW-LUmhm1UVE1ZS2kBaxJ6EFaDU4GUa0
> Content-Length: 72
>
* upload completely sent off: 72 out of 72 bytes
< HTTP/1.1 200 OK
< Access-Control-Allow-Origin: *
< Cache-Control: no-cache, no-store, max-age=0, must-revalidate, value
< Content-Type: application/json; charset=utf-8
< Expires: Thu, 01 Jan 1970 00:00:00 GMT
< Last-Modified: Mon, 06 Dec 2021 16:03:27 GMT
< X-Content-Type-Options: nosniff
< X-Frame-Options: DENY
< X-Request-Id: 9d88e4d1-c888-499f-9ced-34bd6c229bfa
< X-Xss-Protection: 1; mode=block
< Date: Mon, 06 Dec 2021 16:03:27 GMT
< Content-Length: 313
<
* Connection #0 to host iam.api.marmotedu.com left intact
{"metadata":{"id":24,"instanceID":"secret-jpvgjl","name":"secret0","createdAt":"2021-12-07T00:03:27.172+08:00","updatedAt":"2021-12-07T00:03:27.183+08:00"},"username":"admin","secretID":"uFT8uwityjGs7O8LFEO1LfWdgXOzbSu8cSX4","secretKey":"58jxVNRPt7BQswDUkgfhQAsVfPmAASoU","expires":0,"description":"admin secret"}
```

可以看到，请求返回头中返回了X-Request-Id Header，X-Request-Id唯一标识这次 请求。如果这次请求失败，就可以将X-Request-Id提供给运维或者开发，通过XRequest-Id定位出失败的请求，进行排障。另外X-Request-Id在微服务场景中，也可 以透传给其他服务，从而实现请求调用链。

3. 获取 secret0 的详细信息：

```SH
$ curl -XGET -H"Authorization: Bearer ${TOKEN}" http://iam.api.marmotedu.com:8080/v1/secrets/secret0

{"metadata":{"id":24,"instanceID":"secret-jpvgjl","name":"secret0","createdAt":"2021-12-07T00:03:27+08:00","updatedAt":"2021-12-07T00:03:27+08:00"},"username":"admin","secretID":"uFT8uwityjGs7O8LFEO1LfWdgXOzbSu8cSX4","secretKey":"58jxVNRPt7BQswDUkgfhQAsVfPmAASoU","expires":0,"description":"admin secret"}
```

4. 更新 secret0 的描述：

```sh
$ curl -XPUT -H"Authorization: Bearer ${TOKEN}" -d'{"metadata":{"name":"secret"},"expires":0,"description":"admin secret(modify)"}' http://iam.api.marmotedu.com:8080/v1/secrets/secret0

{"metadata":{"id":24,"instanceID":"secret-jpvgjl","name":"secret0","createdAt":"2021-12-07T00:03:27+08:00","updatedAt":"2021-12-07T00:12:58.582+08:00"},"username":"admin","secretID":"uFT8uwityjGs7O8LFEO1LfWdgXOzbSu8cSX4","secretKey":"58jxVNRPt7BQswDUkgfhQAsVfPmAASoU","expires":0,"description":"admin secret(modify)"}
```

5. 获取 secret 列表：

```sh
$ curl -XGET -H"Authorization: Bearer ${TOKEN}" http://iam.api.marmotedu.com:8080/v1/secrets

{"totalCount":2,"items":[{"metadata":{"id":24,"instanceID":"secret-jpvgjl","name":"secret0","createdAt":"2021-12-07T00:03:27+08:00","updatedAt":"2021-12-07T00:12:58+08:00"},"username":"admin","secretID":"uFT8uwityjGs7O8LFEO1LfWdgXOzbSu8cSX4","secretKey":"58jxVNRPt7BQswDUkgfhQAsVfPmAASoU","expires":0,"description":"admin secret(modify)"},{"metadata":{"id":23,"instanceID":"secret-yj8m30","name":"authztest","createdAt":"2021-11-03T00:55:44+08:00","updatedAt":"2021-11-03T00:55:44+08:00"},"username":"admin","secretID":"SuXnTvmGOWu5f95BfonhvYi8uxLBH2y6BOlc","secretKey":"6dF1ENyDWBDGlmR6ipUbUcpkdjgqF5Gh","expires":0,"description":"admin secret"}]}
```

6. 删除 secret0：

```sh
$ curl -XDELETE -H"Authorization: Bearer ${TOKEN}" http://iam.api.marmotedu.com:8080/v1/secrets/secret0

null
```

上面，演示了密钥的使用方法。用户和策略资源类型的使用方法跟密钥类似。详细 的使用方法可以参考 install.sh脚本，该脚本是用来测试 IAM 应用的，里面包含了各个 接口的请求方法。 

##### 测试 IAM 应用的各部分

这里，顺便介绍下如何测试 IAM 应用中的各个部分。确保 iam-apiserver、iam-authz-server、iam-pump 等服务正常运行后，进入到 IAM 项目的根目录，执行以下命 令：

```sh
$ ./scripts/install/test.sh iam::test::test # 测试整个IAM应用是否正常运行
$ ./scripts/install/test.sh iam::test::login # 测试登陆接口是否可以正常访问
$ ./scripts/install/test.sh iam::test::user # 测试用户接口是否可以正常访问
$ ./scripts/install/test.sh iam::test::secret # 测试密钥接口是否可以正常访问
$ ./scripts/install/test.sh iam::test::policy # 测试策略接口是否可以正常访问
$ ./scripts/install/test.sh iam::test::apiserver # 测试iam-apiserver服务是否正常运行
$ ./scripts/install/test.sh iam::test::authz # 测试authz接口是否可以正常访问
$ ./scripts/install/test.sh iam::test::authzserver # 测试iam-authz-server服务是否正常运行
$ ./scripts/install/test.sh iam::test::pump # 测试iam-pump是否正常运行
$ ./scripts/install/test.sh iam::test::iamctl # 测试iamctl工具是否可以正常使用
$ ./scripts/install/test.sh iam::test::man # 测试man文件是否正确安装
```

##### iam-apiserver 的冒烟测试

所以，每次发布完 iam-apiserver 后，可以执行以下命令来完成 iam-apiserver 的冒烟测试：

```sh
$ export IAM_APISERVER_HOST=127.0.0.1 # iam-apiserver部署服务器的IP地址
$ export IAM_APISERVER_INSECURE_BIND_PORT=8080 # iam-apiserver HTTP服务的监听端口
$ ./scripts/install/test.sh iam::test::apiserver
```

### iam-apiserver 代码实现

上面，介绍了 iam-apiserver 的功能和使用方法，这里再来看下 iam-apiserver 具 体的代码实现。从配置处理、启动流程、请求处理流程、代码架构 4 个方面来讲解。 

#### iam-apiserver 配置处理 

iam-apiserver 服务的 main 函数位于 apiserver.go 文件中，可以跟读代码，了解 iam-apiserver 的代码实现。这里，来介绍下 iam-apiserver 服务的一些设计思想。 

首先，来看下 iam-apiserver 中的 3 种配置：Options 配置、应用配置和 HTTP/GRPC 服务配置。

- Options 配置：用来构建命令行参数，它的值来自于命令行选项或者配置文件（也可能 是二者 Merge 后的配置）。Options 可以用来构建应用框架，Options 配置也是应用 配置的输入。 
- 应用配置：iam-apiserver 组件中需要的一切配置。有很多地方需要配置，例如，启动 HTTP/GRPC 需要配置监听地址和端口，初始化数据库需要配置数据库地址、用户名、 密码等。 
- HTTP/GRPC 服务配置：启动 HTTP 服务或者 GRPC 服务需要的配置。

这三种配置的关系如下图：

![image-20211207003149946](IAM-document.assets/image-20211207003149946.png)

Options 配置接管命令行选项，应用配置接管整个应用的配置，HTTP/GRPC 服务配置接 管跟 HTTP/GRPC 服务相关的配置。这 3 种配置独立开来，可以解耦命令行选项、应用和 应用内的服务，使得这 3 个部分可以独立扩展，又不相互影响。 

iam-apiserver 根据 Options 配置来构建命令行参数和应用配置。 

通过 github.com/marmotedu/iam/pkg/app 包(app.go)的 buildCommand 方法来构建命令行参数。这里的核心是，通过 NewApp函数构建 Application 实例时，传入的 Options实现了Flags() (fss cliflag.NamedFlagSets)方法，通过 buildCommand 方法中的以下代码，将 option 的 Flag 添加到 cobra 实例的 FlagSet 中：

```go
if a.options != nil {
		namedFlagSets = a.options.Flags()
		fs := cmd.Flags()
		for _, f := range namedFlagSets.FlagSets {
			fs.AddFlagSet(f)
		}

		...
	}
```

通过 CreateConfigFromOptions函数来构建应用配置：

```go
cfg, err := config.CreateConfigFromOptions(opts)
if err != nil {
	return err
}
```

根据应用配置来构建 HTTP/GRPC 服务配置。例如，以下代码根据应用配置，构建了 HTTP 服务器的 Address 参数：

```go
func (s *InsecureServingOptions) ApplyTo(c *server.Config) error {
  c.InsecureServing = &server.InsecureServingInfo{
  	Address: net.JoinHostPort(s.BindAddress, strconv.Itoa(s.BindPort)),
  }
	return nil
}
```

其中，c *server.Config是 HTTP 服务器的配置，s *InsecureServingOptions是 应用配置。 

#### iam-apiserver 启动流程设计 

接下来，详细看下 iam-apiserver 的启动流程设计。启动流程如下图所示：



困的很，睡觉吧！







