import React, { useMemo } from 'react';
import { Heading, Item as SpectrumItem, ListView } from '@adobe/react-spectrum';
import DragHandle from '@spectrum-icons/workflow/DragHandle';
// TODO @reactima research on React.memo use here
export const FlowEntries = React.memo(function FlowEntries(props) {
    const { list, dragAndDropOptions } = props;
    // 1) Always call hooks unconditionally (before any 'return').
    //    If list or list.items is empty, `tmp` will remain empty.
    const renderList = useMemo(() => {
        const tmp = {};
        if (!list?.items)
            return tmp;
        for (const item of list.items) {
            if (item.parent) {
                if (!tmp[item.parent]) {
                    tmp[item.parent] = { id: item.parent, title: '', children: [] };
                }
                tmp[item.parent].children.push(item);
            }
            else {
                tmp[item.id] = { ...item, children: [] };
            }
        }
        return tmp;
    }, [list]);
    // 2) Now do the early return based on `list.items`.
    if (!list?.items || list.items.length === 0) {
        return <>No FlowEntries or empty</>;
    }
    return (<>
      {Object.values(renderList).map((node) => (<div key={node.id}>
          <Heading level={5}>{node.title}</Heading>
          <ListView key={node.id} density="compact" items={node.children} selectionMode="none" aria-label={`Flow entry ${node.title}`} dragAndDropHooks={dragAndDropOptions}>
            {(child) => (<SpectrumItem textValue={child.title}>
                <DragHandle />
                {child.title}
              </SpectrumItem>)}
          </ListView>
        </div>))}
    </>);
});
//# sourceMappingURL=FlowEntries.jsx.map