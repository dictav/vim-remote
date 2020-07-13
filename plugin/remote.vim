vim9script

import RemoteStart from '../autoload/remote.vim'

command! -nargs=? RemoteStart :call RemoteStart(<f-args>)
