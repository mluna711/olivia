vim.cmd([[
let g:test#strategy = "neovim"
let g:test#neovim#start_normal = 1
]])
return {
	"vim-test/vim-test",
	event = "VeryLazy",
	config = function()
	end,
}
