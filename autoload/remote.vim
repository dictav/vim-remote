vim9script

const cmd = expand('<sfile>:p:h:h') .. '/bin/vimremote'

if !exists('g:vimremote_log')
  g:vimremote_log = 0
endif

if !exists('g:vimremote_log_file')
  g:vimremote_log_file = 'vimremote.log'
endif

if !exists('g:vimremote_default_sock_file')
  g:vimremote_default_sock_file = 'vimremote.sock'
endif

def s:callback(ch: channel, msg: string)
enddef

export def RemoteStart(sock = g:vimremote_default_sock_file)
  let jobcmd = [cmd, '-serve', '-sock', sock]
  if g:vimremote_log 
    add(jobcmd, '-log')
    add(jobcmd, g:vimremote_log_file)
  endif

  let job = job_start(jobcmd, {
    'mode': 'json',
    'callback': function('s:callback')
  })
enddef
