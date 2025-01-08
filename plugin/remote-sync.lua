local chan

local function ensure_job()
  if chan then
    return chan
  end
  chan = vim.fn.jobstart({ 'remotesync' }, { rpc = true })
  return chan
end

vim.api.nvim_create_user_command('SelectConnection', function(args)
  vim.fn.rpcrequest(ensure_job(), 'selectConnection', args.fargs)
end, { nargs = '*' })
