

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



## 设计模式之GoF

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



## Go编码规范

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

















