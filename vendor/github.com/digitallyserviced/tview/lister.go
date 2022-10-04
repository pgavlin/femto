package tview

import (
	// "fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ListItemsVisibility uint16

const (
	ListItemVisible ListItemsVisibility = iota << 1
	ListItemNotVisible
	ListItemHidden
)

// listerItem represents one item in a List.
type listerItem struct {
	selected      func()
	mainText      string
	secondaryText string
	shortcut      rune
}

type ListItemSelected interface {
	Selected(idx int, i interface{}, lis []*ListItem)
	Changed(idx int, selected bool, i interface{}, lis []*ListItem)
}

type ListItemText interface {
	MainText() string
	SecondaryText() string
	Shortcut() rune
}

type ListStyles struct {
	main, sec, short, sel string
}

type ListItemDrawable interface {
	GetPrimitive() Primitive
}

type ListItemVisibility interface {
	Visibility() ListItemsVisibility
}

type ListItem interface {
	ListItemText
	ListItemSelected
	ListItemVisibility
}

func (li listerItem) Visibility() ListItemsVisibility {
	return ListItemVisible
}

func (li listerItem) MainText() string {
	return li.mainText
}

func (li listerItem) SecondaryText() string {
	return li.secondaryText
}

func (li listerItem) Shortcut() rune {
	return li.shortcut
}

func (li listerItem) Selected(idx int, i interface{}, lis []*ListItem) {
}

func (li listerItem) Changed(idx int, selected bool, i interface{}, lis []*ListItem) {
}

// Lister displays rows of items, each of which can be selected.
//
// See https://github.com/rivo/tview/wiki/Lister for an example.
type Lister struct {
	selectedStyle      tcell.Style
	mainTextStyle      tcell.Style
	secondaryTextStyle tcell.Style
	shortcutStyle      tcell.Style
	selected           func(index int, mainText, secondaryText string, shortcut rune)
	changed            func(index int, mainText, secondaryText string, shortcut rune)
	*Box
	done              func()
	itemsLister       func() []*ListItem
	items             []*ListItem
	currentItem       int
	itemOffset        int
	horizontalOffset  int
	wrapAround        bool
	overflowing       bool
	showSecondaryText bool
	highlightFullLine bool
	selectedFocusOnly bool
}

// NewList returns a new form.
func NewLister() *Lister {
	return &Lister{
		Box:                NewBox(),
		showSecondaryText:  true,
		wrapAround:         true,
		mainTextStyle:      tcell.StyleDefault.Foreground(Styles.PrimaryTextColor),
		secondaryTextStyle: tcell.StyleDefault.Foreground(Styles.TertiaryTextColor),
		shortcutStyle:      tcell.StyleDefault.Foreground(Styles.SecondaryTextColor),
		selectedStyle:      tcell.StyleDefault.Foreground(Styles.PrimitiveBackgroundColor).Background(Styles.PrimaryTextColor),
	}
}

// SetCurrentItem sets the currently selected item by its index, starting at 0
// for the first item. If a negative index is provided, items are referred to
// from the back (-1 = last item, -2 = second-to-last item, and so on). Out of
// range indices are clamped to the beginning/end.
//
// Calling this function triggers a "changed" event if the selection changes.
func (l *Lister) SetCurrentItem(index int) *Lister {
	if index < 0 {
		index = len(l.items) + index
	}
	if index >= len(l.items) {
		index = len(l.items) - 1
	}
	if index < 0 {
		index = 0
	}

	if index != l.currentItem && l.changed != nil {
		item := *l.items[index]
		l.changed(index, item.MainText(), item.SecondaryText(), item.Shortcut())
	}

	l.currentItem = index

	return l
}

// SetCurrent sets the item based on a reference to the actual item instead of index
// ranges through the items to find passed reference and updates the currentItem
// index value

func (l *Lister) SetCurrent(i interface{}) *Lister {
	// if index < 0 {
	// 	index = len(l.items) + index
	// }
	// if index >= len(l.items) {
	// 	index = len(l.items) - 1
	// }
	// if index < 0 {
	// 	index = 0
	// }
	//
	// if index != l.currentItem && l.changed != nil {
	// 	item := l.items[index]
	// 	l.changed(index, item.MainText(), item.SecondaryText(), item.Shortcut())
	// }
	index := -1
	for num, v := range l.items {
		if v == i {
			index = num
		}
	}

	l.currentItem = index

	return l
}

// GetCurrentItem returns the index of the currently selected list item,
// starting at 0 for the first item.
func (l *Lister) GetCurrentItem() int {
	return l.currentItem
}

// SetOffset sets the number of items to be skipped (vertically) as well as the
// number of cells skipped horizontally when the list is drawn. Note that one
// item corresponds to two rows when there are secondary texts. Shortcut()s are
// always drawn.
//
// These values may change when the list is drawn to ensure the currently
// selected item is visible and item texts move out of view. Users can also
// modify these values by interacting with the list.
func (l *Lister) SetOffset(items, horizontal int) *Lister {
	l.itemOffset = items
	l.horizontalOffset = horizontal
	return l
}

// GetOffset returns the number of items skipped while drawing, as well as the
// number of cells item text is moved to the left. See also SetOffset() for more
// information on these values.
func (l *Lister) GetOffset() (int, int) {
	return l.itemOffset, l.horizontalOffset
}

// RemoveItem removes the item with the given index (starting at 0) from the
// list. If a negative index is provided, items are referred to from the back
// (-1 = last item, -2 = second-to-last item, and so on). Out of range indices
// are clamped to the beginning/end, i.e. unless the list is empty, an item is
// always removed.
//
// The currently selected item is shifted accordingly. If it is the one that is
// removed, a "changed" event is fired.
func (l *Lister) RemoveItem(index int) *Lister {
	if len(l.items) == 0 {
		return l
	}

	// Adjust index.
	if index < 0 {
		index = len(l.items) + index
	}
	if index >= len(l.items) {
		index = len(l.items) - 1
	}
	if index < 0 {
		index = 0
	}

	// Remove item.
	l.items = append(l.items[:index], l.items[index+1:]...)

	// If there is nothing left, we're done.
	if len(l.items) == 0 {
		return l
	}

	// Shift current item.
	previousCurrentItem := l.currentItem
	if l.currentItem >= index {
		l.currentItem--
	}

	// Fire "changed" event for removed items.
	if previousCurrentItem == index && l.changed != nil {
		item := *l.items[l.currentItem]
		l.changed(l.currentItem, item.MainText(), item.SecondaryText(), item.Shortcut())
	}

	return l
}

// SetMainText()Color sets the color of the items' main text.
func (l *Lister) SetMainTextColor(color tcell.Color) *Lister {
	l.mainTextStyle = l.mainTextStyle.Foreground(color)
	return l
}

// SetMainText()Style sets the style of the items' main text. Note that the
// background color is ignored in order not to override the background color of
// the list itself.
func (l *Lister) SetMainTextStyle(style tcell.Style) *Lister {
	l.mainTextStyle = style
	return l
}

// SetSecondaryText()Color sets the color of the items' secondary text.
func (l *Lister) SetSecondaryTextColor(color tcell.Color) *Lister {
	l.secondaryTextStyle = l.secondaryTextStyle.Foreground(color)
	return l
}

// SetSecondaryText()Style sets the style of the items' secondary text. Note that
// the background color is ignored in order not to override the background color
// of the list itself.
func (l *Lister) SetSecondaryTextStyle(style tcell.Style) *Lister {
	l.secondaryTextStyle = style
	return l
}

// SetShortcut()Color sets the color of the items' shortcut.
func (l *Lister) SetShortcutColor(color tcell.Color) *Lister {
	l.shortcutStyle = l.shortcutStyle.Foreground(color)
	return l
}

// SetShortcut()Style sets the style of the items' shortcut. Note that the
// background color is ignored in order not to override the background color of
// the list itself.
func (l *Lister) SetShortcutStyle(style tcell.Style) *Lister {
	l.shortcutStyle = style
	return l
}

// SetSelectedTextColor sets the text color of selected items. Note that the
// color of main text characters that are different from the main text color
// (e.g. color tags) is maintained.
func (l *Lister) SetSelectedTextColor(color tcell.Color) *Lister {
	l.selectedStyle = l.selectedStyle.Foreground(color)
	return l
}

// SetSelectedBackgroundColor sets the background color of selected items.
func (l *Lister) SetSelectedBackgroundColor(color tcell.Color) *Lister {
	l.selectedStyle = l.selectedStyle.Background(color)
	return l
}

// SetSelectedStyle sets the style of the selected items. Note that the color of
// main text characters that are different from the main text color (e.g. color
// tags) is maintained.
func (l *Lister) SetSelectedStyle(style tcell.Style) *Lister {
	l.selectedStyle = style
	return l
}

// SetSelectedFocusOnly sets a flag which determines when the currently selected
// list item is highlighted. If set to true, selected items are only highlighted
// when the list has focus. If set to false, they are always highlighted.
func (l *Lister) SetSelectedFocusOnly(focusOnly bool) *Lister {
	l.selectedFocusOnly = focusOnly
	return l
}

// SetHighlightFullLine sets a flag which determines whether the colored
// background of selected items spans the entire width of the view. If set to
// true, the highlight spans the entire view. If set to false, only the text of
// the selected item from beginning to end is highlighted.
func (l *Lister) SetHighlightFullLine(highlight bool) *Lister {
	l.highlightFullLine = highlight
	return l
}

// ShowSecondaryText() determines whether or not to show secondary item texts.
func (l *Lister) ShowSecondaryText(show bool) *Lister {
	l.showSecondaryText = show
	return l
}

// SetWrapAround sets the flag that determines whether navigating the list will
// wrap around. That is, navigating downwards on the last item will move the
// selection to the first item (similarly in the other direction). If set to
// false, the selection won't change when navigating downwards on the last item
// or navigating upwards on the first item.
func (l *Lister) SetWrapAround(wrapAround bool) *Lister {
	l.wrapAround = wrapAround
	return l
}

// SetChangedFunc sets the function which is called when the user navigates to
// a list item. The function receives the item's index in the list of items
// (starting with 0), its main text, secondary text, and its shortcut rune.
//
// This function is also called when the first item is added or when
// SetCurrentItem() is called.
func (l *Lister) SetChangedFunc(handler func(index int, mainText string, secondaryText string, shortcut rune)) *Lister {
	l.changed = handler
	return l
}

// SetSelectedFunc sets the function which is called when the user selects a
// list item by pressing Enter on the current selection. The function receives
// the item's index in the list of items (starting with 0), its main text,
// secondary text, and its shortcut rune.
func (l *Lister) SetSelectedFunc(handler func(int, string, string, rune)) *Lister {
	l.selected = handler
	return l
}

// SetDoneFunc sets a function which is called when the user presses the Escape
// key.
func (l *Lister) SetDoneFunc(handler func()) *Lister {
	l.done = handler
	return l
}

// AddItem calls InsertItem() with an index of -1.
func (l *Lister) AddItem(mainText, secondaryText string, shortcut rune, selected func()) *Lister {
	l.InsertItem(-1, mainText, secondaryText, shortcut, selected)
	return l
}

// InsertItem adds a new item to the list at the specified index. An index of 0
// will insert the item at the beginning, an index of 1 before the second item,
// and so on. An index of GetItemCount() or higher will insert the item at the
// end of the list. Negative indices are also allowed: An index of -1 will
// insert the item at the end of the list, an index of -2 before the last item,
// and so on. An index of -GetItemCount()-1 or lower will insert the item at the
// beginning.
//
// An item has a main text which will be highlighted when selected. It also has
// a secondary text which is shown underneath the main text (if it is set to
// visible) but which may remain empty.
//
// The shortcut is a key binding. If the specified rune is entered, the item
// is selected immediately. Set to 0 for no binding.
//
// The "selected" callback will be invoked when the user selects the item. You
// may provide nil if no such callback is needed or if all events are handled
// through the selected callback set with SetSelectedFunc().
//
// The currently selected item will shift its position accordingly. If the list
// was previously empty, a "changed" event is fired because the new item becomes
// selected.
func (l *Lister) InsertItem(index int, mainText, secondaryText string, shortcut rune, selected func()) *Lister {
	item := &listerItem{
		mainText:      mainText,
		secondaryText: secondaryText,
		shortcut:      shortcut,
		selected:      selected,
	}

	// Shift index to range.
	if index < 0 {
		index = len(l.items) + index + 1
	}
	if index < 0 {
		index = 0
	} else if index > len(l.items) {
		index = len(l.items)
	}

	// Shift current item.
	if l.currentItem < len(l.items) && l.currentItem >= index {
		l.currentItem++
	}

	// Insert item (make space for the new item, then shift and insert).
	l.items = append(l.items, nil)
	if index < len(l.items)-1 { // -1 because l.items has already grown by one item.
		copy(l.items[index+1:], l.items[index:])
	}

	// var litem *ListItem = item.(*ListItem)

	litem := ListItem(*item)
	l.items[index] = &litem
	// l.items[index] = ListItem(*item)

	// Fire a "change" event for the first item in the list.
	if len(l.items) == 1 && l.changed != nil {
		item := *l.items[0]
		l.changed(0, item.MainText(), item.SecondaryText(), item.Shortcut())
	}

	return l
}

// GetItem returns the ListItem at the specified index in the list.
func (l *Lister) GetItem(index int) *ListItem {
	if index < 0 {
		index = len(l.items) + index
	}
	if index >= len(l.items) {
		index = index - len(l.items) - 1
		return l.GetItem(index)
	}
	if index < 0 {
		index = 0
	}

	var item ListItem = (*l.items[index])
	// }

	// l.currentItem = index

	return &item
}

// GetItemCount returns the number of items in the list.
func (l *Lister) GetItemCount() int {
	return len(l.items)
}

// GetItemText returns an item's texts (main and secondary). Panics if the index
// is out of range.
func (l *Lister) GetItemText(index int) (main, secondary string) {
	return (*l.items[index]).MainText(), (*l.items[index]).SecondaryText()
}

// SetItemText sets an item's main and secondary text. Panics if the index is
// out of range.
// func (l *Lister) SetItemText(index int, main, secondary string) *Lister {
// 	item := l.items[index]
// 	item.MainText() = main
// 	item.SecondaryText() = secondary
// 	return l
// }

// FindItems searches the main and secondary texts for the given strings and
// returns a list of item indices in which those strings are found. One of the
// two search strings may be empty, it will then be ignored. Indices are always
// returned in ascending order.
//
// If mustContainBoth is set to true, mainSearch must be contained in the main
// text AND secondarySearch must be contained in the secondary text. If it is
// false, only one of the two search strings must be contained.
//
// Set ignoreCase to true for case-insensitive search.
func (l *Lister) FindItems(mainSearch, secondarySearch string, mustContainBoth, ignoreCase bool) (indices []int) {
	if mainSearch == "" && secondarySearch == "" {
		return
	}

	if ignoreCase {
		mainSearch = strings.ToLower(mainSearch)
		secondarySearch = strings.ToLower(secondarySearch)
	}

	for index, itemP := range l.items {
		item := *itemP
		mainText := item.MainText()
		secondaryText := item.SecondaryText()
		if ignoreCase {
			mainText = strings.ToLower(mainText)
			secondaryText = strings.ToLower(secondaryText)
		}

		// strings.Contains() always returns true for a "" search.
		mainContained := strings.Contains(mainText, mainSearch)
		secondaryContained := strings.Contains(secondaryText, secondarySearch)
		if mustContainBoth && mainContained && secondaryContained ||
			!mustContainBoth && (mainText != "" && mainContained || secondaryText != "" && secondaryContained) {
			indices = append(indices, index)
		}
	}

	return
}

func (f *Lister) SetItemLister(il func() []*ListItem) {
	f.itemsLister = il
}

func (f *Lister) UpdateListItems() {
  if f.itemsLister != nil {
    f.items = f.itemsLister()
  }
}

func (f *Lister) SetListItems(li []*ListItem) {
	f.items = li
}

// Clear removes all items from the list.
func (l *Lister) ClearItems() *Lister {
	l.items = nil
	return l
}

// Clear removes all items from the list.
func (l *Lister) Clear() *Lister {
	l.items = nil
	l.currentItem = 0
	return l
}

// Draw draws this primitive onto the screen.
func (l *Lister) Draw(screen tcell.Screen) {
	l.DrawForSubclass(screen, l)

	// Determine the dimensions.
	x, y, width, height := l.GetInnerRect()
	bottomLimit := y + height
	_, totalHeight := screen.Size()
	if bottomLimit > totalHeight {
		bottomLimit = totalHeight
	}

	// Do we show any shortcuts?
	// var showShortcuts bool
	// for _, itemP := range l.items {
	//    item := *itemP
	// 	if item.Shortcut() != 0 {
	// 		showShortcuts = true
	// 		x += 4
	// 		width -= 4
	// 		break
	// 	}
	// }

	// Adjust offset to keep the current selection in view.
	if l.currentItem < l.itemOffset {
		l.itemOffset = l.currentItem
	} else if l.showSecondaryText {
		if 2*(l.currentItem-l.itemOffset) >= height-1 {
			l.itemOffset = (2*l.currentItem + 3 - height) / 2
		}
	} else {
		if l.currentItem-l.itemOffset >= height {
			l.itemOffset = l.currentItem + 1 - height
		}
	}
	if l.horizontalOffset < 0 {
		l.horizontalOffset = 0
	}

	// Draw the list items.
	var (
		maxWidth    int  // The maximum printed item width.
		overflowing bool // Whether a text's end exceeds the right border.
	)
	// k, v := range l.items

	si := 0
	for y < bottomLimit {
		// index, itemP := range l.items
		index := si
		item := (*l.GetItem(index))
		// itemP := l.items[index]
		// item := (*itemP)
		if index < l.itemOffset {
			continue
		}

		if y >= bottomLimit {
			break
		}

		// Shortcut()s.
		// if showShortcuts && item.Shortcut() != 0 {
		// 	printWithStyle(screen, fmt.Sprintf("(%s)", string(item.Shortcut())), x-5, y, 0, 4, AlignRight, l.shortcutStyle, true)
		// }

		// Main text.
		_, printedWidth, _, end := printWithStyle(screen, item.MainText(), x, y, l.horizontalOffset, width, AlignLeft, l.mainTextStyle, true)
		if printedWidth > maxWidth {
			maxWidth = printedWidth
		}
		if end < len(item.MainText()) {
			overflowing = true
		}

		// Background color of selected text.
		if index == l.currentItem && (!l.selectedFocusOnly || l.HasFocus()) {
			textWidth := width
			if !l.highlightFullLine {
				if w := TaggedStringWidth(item.MainText()); w < textWidth {
					textWidth = w
				}
			}

			mainTextColor, _, _ := l.mainTextStyle.Decompose()
			for bx := 0; bx < textWidth; bx++ {
				m, c, style, _ := screen.GetContent(x+bx, y)
				fg, _, _ := style.Decompose()
				style = l.selectedStyle
				if fg != mainTextColor {
					style = style.Foreground(fg)
				}
				screen.SetContent(x+bx, y, m, c, style)
			}
		}

		y++

		if y >= bottomLimit {
			break
		}

		// Secondary text.
		if l.showSecondaryText {
			_, printedWidth, _, end := printWithStyle(screen, item.SecondaryText(), x, y, l.horizontalOffset, width, AlignLeft, l.secondaryTextStyle, true)
			if printedWidth > maxWidth {
				maxWidth = printedWidth
			}
			if end < len(item.SecondaryText()) {
				overflowing = true
			}
			y++
		}
	}

	// We don't want the item text to get out of view. If the horizontal offset
	// is too high, we reset it and redraw. (That should be about as efficient
	// as calculating everything up front.)
	if l.horizontalOffset > 0 && maxWidth < width {
		l.horizontalOffset -= width - maxWidth
		l.Draw(screen)
	}
	l.overflowing = overflowing
}

// InputHandler returns the handler for this primitive.
func (l *Lister) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if event.Key() == tcell.KeyEscape {
			if l.done != nil {
				l.done()
			}
			return
		} else if len(l.items) == 0 {
			return
		}

		previousItem := l.currentItem

		switch key := event.Key(); key {
		case tcell.KeyTab, tcell.KeyDown:
			l.currentItem++
		case tcell.KeyBacktab, tcell.KeyUp:
			l.currentItem--
		case tcell.KeyRight:
			if l.overflowing {
				l.horizontalOffset += 2 // We shift by 2 to account for two-cell characters.
			} else {
				l.currentItem++
			}
		case tcell.KeyLeft:
			if l.horizontalOffset > 0 {
				l.horizontalOffset -= 2
			} else {
				l.currentItem--
			}
		case tcell.KeyHome:
			l.currentItem = 0
		case tcell.KeyEnd:
			l.currentItem = len(l.items) - 1
		case tcell.KeyPgDn:
			_, _, _, height := l.GetInnerRect()
			l.currentItem += height
			if l.currentItem >= len(l.items) {
				l.currentItem = len(l.items) - 1
			}
		case tcell.KeyPgUp:
			_, _, _, height := l.GetInnerRect()
			l.currentItem -= height
			if l.currentItem < 0 {
				l.currentItem = 0
			}
		case tcell.KeyEnter:
			if l.currentItem >= 0 && l.currentItem < len(l.items) {
				item := *l.items[l.currentItem]
				item.Selected(l.currentItem, item, l.items)
				if l.selected != nil {
					l.selected(l.currentItem, item.MainText(), item.SecondaryText(), item.Shortcut())
				}
			}
		case tcell.KeyRune:
			ch := event.Rune()
			if ch != ' ' {
				// It's not a space bar. Is it a shortcut?
				var found bool
				for index, itemP := range l.items {
					item := (*itemP)
					if item.Shortcut() == ch {
						// We have a shortcut.
						found = true
						l.currentItem = index
						break
					}
				}
				if !found {
					break
				}
			}
			item := *l.items[l.currentItem]
			item.Selected(l.currentItem, item, l.items)
			// if item.Selected != nil {
			// item.Selected()
			// }
			if l.selected != nil {
				l.selected(l.currentItem, item.MainText(), item.SecondaryText(), item.Shortcut())
			}
		}

		if l.currentItem < 0 {
			if l.wrapAround {
				l.currentItem = len(l.items) - 1
			} else {
				l.currentItem = 0
			}
		} else if l.currentItem >= len(l.items) {
			if l.wrapAround {
				l.currentItem = 0
			} else {
				l.currentItem = len(l.items) - 1
			}
		}

		if l.currentItem != previousItem && l.currentItem < len(l.items) && l.changed != nil {
			item := *l.items[l.currentItem]
			l.changed(l.currentItem, item.MainText(), item.SecondaryText(), item.Shortcut())
		}
	})
}

// indexAtPoint returns the index of the list item found at the given position
// or a negative value if there is no such list item.
func (l *Lister) indexAtPoint(x, y int) int {
	rectX, rectY, width, height := l.GetInnerRect()
	if rectX < 0 || rectX >= rectX+width || y < rectY || y >= rectY+height {
		return -1
	}

	index := y - rectY
	if l.showSecondaryText {
		index /= 2
	}
	index += l.itemOffset

	if index >= len(l.items) {
		return -1
	}
	return index
}

// MouseHandler returns the mouse handler for this primitive.
func (l *Lister) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return l.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if !l.InRect(event.Position()) {
			return false, nil
		}

		// Process mouse event.
		switch action {
		case MouseLeftClick:
			setFocus(l)
			index := l.indexAtPoint(event.Position())
			if index != -1 {
				item := *l.items[index]
				item.Selected(index, item, l.items)
				// item.Selected
				if l.selected != nil {
					l.selected(index, item.MainText(), item.SecondaryText(), item.Shortcut())
				}
				if index != l.currentItem && l.changed != nil {
					l.changed(index, item.MainText(), item.SecondaryText(), item.Shortcut())
				}
				l.currentItem = index
			}
			consumed = true
		case MouseScrollUp:
			if l.itemOffset > 0 {
				l.itemOffset--
			}
			consumed = true
		case MouseScrollDown:
			lines := len(l.items) - l.itemOffset
			if l.showSecondaryText {
				lines *= 2
			}
			if _, _, _, height := l.GetInnerRect(); lines > height {
				l.itemOffset++
			}
			consumed = true
		}

		return
	})
}
